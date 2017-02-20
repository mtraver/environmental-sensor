"""Log data to a Google Sheets spreadsheet using service account credentials."""
import httplib2

import apiclient
from oauth2client.service_account import ServiceAccountCredentials

SCOPE = 'https://www.googleapis.com/auth/spreadsheets'
DISCOVERY_URL = 'https://sheets.googleapis.com/$discovery/rest?version=v4'


def append_to_sheet(keyfile, spreadsheet_id, values, sheet='Sheet1'):
  credentials = ServiceAccountCredentials.from_json_keyfile_name(
      keyfile, scopes=SCOPE)

  http = credentials.authorize(httplib2.Http())
  service = apiclient.discovery.build('sheets', 'v4', http=http,
                                      discoveryServiceUrl=DISCOVERY_URL)

  service.spreadsheets().values().append(spreadsheetId=spreadsheet_id,
    range=sheet, valueInputOption='RAW',
    body={'values': values}).execute()
