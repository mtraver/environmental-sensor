"""Read and write Google Sheets using service account credentials."""
import apiclient
import httplib2
from oauth2client.service_account import ServiceAccountCredentials
import pandas

SCOPE = 'https://www.googleapis.com/auth/spreadsheets'
DISCOVERY_URL = 'https://sheets.googleapis.com/$discovery/rest?version=v4'

DEFAULT_SHEET_RANGE = 'Sheet1'


def _get_authenticated_service(keyfile):
  credentials = ServiceAccountCredentials.from_json_keyfile_name(
      keyfile, scopes=SCOPE)

  http = credentials.authorize(httplib2.Http())
  service = apiclient.discovery.build('sheets', 'v4', http=http,
                                      discoveryServiceUrl=DISCOVERY_URL)

  return service


def append_to_sheet(keyfile, spreadsheet_id, values,
                    sheet_range=DEFAULT_SHEET_RANGE):
  """Appends data to a Google Sheet.

  Args:
    keyfile: Path to service account JSON keyfile.
    spreadsheet_id: ID of Google Sheet, from its URL.
    values: List of values to append to the sheet, one element per column.
    sheet_range: A range, in A1 notation, specifying a "table" in the
      spreadsheet to which to append data. Defaults to 'Sheet1'.
      See these pages for details on how this works:
      https://developers.google.com/sheets/api/guides/concepts#a1_notation
      https://developers.google.com/sheets/api/guides/values#appending_values

  """
  service = _get_authenticated_service(keyfile)

  service.spreadsheets().values().append(
      spreadsheetId=spreadsheet_id, range=sheet_range, valueInputOption='RAW',
      body={'values': values}).execute()


def read_sheet(keyfile, spreadsheet_id, sheet_range=DEFAULT_SHEET_RANGE,
               header=0, index_col=None, parse_dates=None):
  """Reads data from a Google Sheet into a pandas.DataFrame.

  Args:
    keyfile: Path to service account JSON keyfile.
    spreadsheet_id: ID of Google Sheet, from its URL.
    sheet_range: Range to read, in A1 notation. Defaults to 'Sheet1'.
      See this page for details:
      https://developers.google.com/sheets/api/guides/concepts#a1_notation
    header: The row to use as column headers. Defaults to 0.
      If None, then no headers are set.
    index_col: The name of the column to use as the index. Defaults to None,
      in which case pandas' default numeric index is used.
    parse_dates: List of column names to attempt to parse as datetimes.

  Returns:
    A pandas.DataFrame
  """
  service = _get_authenticated_service(keyfile)

  response = service.spreadsheets().values().get(
      spreadsheetId=spreadsheet_id, majorDimension='ROWS',
      range=sheet_range).execute()
  raw_data = response['values']

  if header is not None:
    headers = raw_data.pop(header)
    data = pandas.DataFrame(raw_data, columns=headers)
  else:
    data = pandas.DataFrame(raw_data)

  if parse_dates is not None:
    for col in parse_dates:
      data[col] = pandas.to_datetime(data[col], infer_datetime_format=True)

  if index_col is not None:
    data.set_index(index_col, inplace=True)

  return data
