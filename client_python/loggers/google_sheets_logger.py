"""Read and write Google Sheets using service account credentials."""
import apiclient
import httplib2
from oauth2client.service_account import ServiceAccountCredentials

import loggers.base_logger

DEFAULT_SHEET_RANGE = 'Sheet1'


class BaseSheetsLogger(loggers.base_logger.Logger):
  """Base class for classes that read and write Google Sheets."""

  SCOPE = 'https://www.googleapis.com/auth/spreadsheets'
  DISCOVERY_URL = 'https://sheets.googleapis.com/$discovery/rest?version=v4'

  def __init__(self, keyfile, spreadsheet_id, sheet_range=DEFAULT_SHEET_RANGE):
    """Creates a Google Sheets data logger.

    Args:
      keyfile: Path to service account JSON keyfile.
      spreadsheet_id: ID of Google Sheet, from its URL.
      sheet_range: A range, in A1 notation, specifying a "table" in the
                   spreadsheet to which to append data or from which to read
                   data. Defaults to 'Sheet1'.
                   See these pages for details on how this works:
                   https://developers.google.com/sheets/api/guides/concepts#a1_notation
                   https://developers.google.com/sheets/api/guides/values#appending_values
    """
    self.keyfile = keyfile
    self.spreadsheet_id = spreadsheet_id
    self.sheet_range = sheet_range

  def _get_authenticated_service(self):
    credentials = ServiceAccountCredentials.from_json_keyfile_name(
        self.keyfile, scopes=self.SCOPE)

    http = credentials.authorize(httplib2.Http())
    service = apiclient.discovery.build('sheets', 'v4', http=http,
                                        discoveryServiceUrl=self.DISCOVERY_URL)

    return service


class SheetsLogger(BaseSheetsLogger):

  def log(self, timestamp, values):
    """Appends data to a Google Sheet.

    Args:
      timestamp: A datetime.datetime.
      values: List of values to append to the sheet, one element per column.
    """
    service = self._get_authenticated_service()

    # The Sheets API expects a list of lists, with each inner list representing
    # a major dimension (which in this case is rows). Put the timestamp in the
    # first column, followed by the data.
    values_to_log = [[timestamp.isoformat()] + values]

    service.spreadsheets().values().append(
        spreadsheetId=self.spreadsheet_id, range=self.sheet_range,
        valueInputOption='RAW', body={'values': values_to_log}).execute()


class SheetsReader(BaseSheetsLogger):

  def read(self):
    """Reads data from a Google Sheet.

    Returns:
      A list, where each element is a list of the cell values
      of a single row of the Google Sheet.
    """
    service = self._get_authenticated_service()

    response = service.spreadsheets().values().get(
        spreadsheetId=self.spreadsheet_id, majorDimension='ROWS',
        range=self.sheet_range).execute()

    return response['values']
