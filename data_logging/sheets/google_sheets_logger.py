"""Read and write Google Sheets using service account credentials."""
import apiclient
import httplib2
from oauth2client.service_account import ServiceAccountCredentials

SCOPE = 'https://www.googleapis.com/auth/spreadsheets'
DISCOVERY_URL = 'https://sheets.googleapis.com/$discovery/rest?version=v4'

DEFAULT_SHEET_RANGE = 'Sheet1'


class _SheetsLogger(object):
  """Base class for classes that read and write Google Sheets."""

  def __init__(self, keyfile, spreadsheet_id):
    """Creates a Google Sheets data logger.

    Args:
      keyfile: Path to service account JSON keyfile.
      spreadsheet_id: ID of Google Sheet, from its URL.
    """
    self.keyfile = keyfile
    self.spreadsheet_id = spreadsheet_id

  def _get_authenticated_service(self):
    credentials = ServiceAccountCredentials.from_json_keyfile_name(
        self.keyfile, scopes=SCOPE)

    http = credentials.authorize(httplib2.Http())
    service = apiclient.discovery.build('sheets', 'v4', http=http,
                                        discoveryServiceUrl=DISCOVERY_URL)

    return service


class Writer(_SheetsLogger):

  def append(self, timestamp, values, sheet_range=DEFAULT_SHEET_RANGE):
    """Appends data to a Google Sheet.

    Args:
      timestamp: A datetime.datetime.
      values: List of values to append to the sheet, one element per column.
      sheet_range: A range, in A1 notation, specifying a "table" in the
                   spreadsheet to which to append data. Defaults to 'Sheet1'.
                   See these pages for details on how this works:
                   https://developers.google.com/sheets/api/guides/concepts#a1_notation
                   https://developers.google.com/sheets/api/guides/values#appending_values

    """
    service = self._get_authenticated_service()

    # The Sheets API expects a list of lists, with each inner list representing
    # a major dimension (which in this case is rows). Put the timestamp in the
    # first column, followed by the data.
    values_to_log = [[timestamp.isoformat()] + values]

    service.spreadsheets().values().append(
        spreadsheetId=self.spreadsheet_id, range=sheet_range,
        valueInputOption='RAW', body={'values': values_to_log}).execute()


class Reader(_SheetsLogger):

  def read(self, sheet_range=DEFAULT_SHEET_RANGE):
    """Reads data from a Google Sheet.

    Args:
      sheet_range: Range to read, in A1 notation. Defaults to 'Sheet1'.
                   See this page for details:
                   https://developers.google.com/sheets/api/guides/concepts#a1_notation

    Returns:
      A list, where each element is a list of the cell values
      of a single row of the Google Sheet.
    """
    service = self._get_authenticated_service()

    response = service.spreadsheets().values().get(
        spreadsheetId=self.spreadsheet_id, majorDimension='ROWS',
        range=sheet_range).execute()

    return response['values']
