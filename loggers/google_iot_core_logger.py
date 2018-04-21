"""Log to Google Cloud IoT Core."""
import base64
import datetime
import json
import os
import ssl

import jwt
import paho.mqtt.client as mqtt
import requests

import loggers.base_logger


SUPPORTED_ALGORITHMS = ['RS256', 'ES256']
DEFAULT_ALG = 'ES256'


class CloudIotLogger(loggers.base_logger.GCPLogger):
  """Base class for loggers that log to Google Cloud IoT Core."""

  DEFAULT_CLOUD_REGION = 'us-central1'

  def __init__(self, project_id, registry_id, priv_key_file, device_id,
               signing_alg, cloud_region):
    """Creates a logger that logs data via Google Cloud IoT Core.

    Args:
      project_id: The ID of the Google Cloud project the device belongs to.
      registry_id: The Cloud IoT Core device registry this device belongs to.
      priv_key_file: Path to a file containing the device's RSA256 or ES256
                     private key. It must have been added to the IoT Core
                     registry.
      device_id: Device ID string. Must be the same as the device ID registered
                 with Cloud IoT Core.
      signing_alg: The signing algorithm used to make a JWT for authentication
                   with Cloud IoT Core. Must be one of {RSA256, ES256}.
      cloud_region: The cloud region of the Cloud IoT Core device registry.

    Raises:
      ValueError: If signing_alg is not one of {RSA256, ES256}.
    """
    super(CloudIotLogger, self).__init__(project_id, device_id)

    if signing_alg not in SUPPORTED_ALGORITHMS:
      raise ValueError('Algorithm not supported: "{}"'.format(signing_alg))
    self._signing_alg = signing_alg

    self._registry_id = registry_id
    self._priv_key_file = priv_key_file
    self._cloud_region = cloud_region

  def _create_jwt(self, validity_minutes=60):
    """Creates a JWT (https://jwt.io) to establish an MQTT connection.

    Args:
     validity_minutes: The number of minutes for which the JWT is valid.

    Returns:
      A JWT generated from this instance's project_id and private key, which
      expires in validity_minutes minutes. After validity_minutes minutes, your
      client will be disconnected, and a new JWT will have to be generated.

    Raises:
      ValueError: If the instance's private key file does not
                  contain a known key.
    """
    token = {
        # The time at which the token was issued
        'iat': datetime.datetime.utcnow(),
        # The time the token expires
        'exp': datetime.datetime.utcnow() + datetime.timedelta(
            minutes=validity_minutes),
        # The audience field should always be set to the GCP project ID
        'aud': self._project_id
    }

    with open(self._priv_key_file, 'r') as f:
      private_key = f.read()

    return jwt.encode(token, private_key, algorithm=self._signing_alg)


def error_str(return_code):
  """Convert a Paho error to a human readable string."""
  return '{}: {}'.format(return_code, mqtt.error_string(return_code))


def on_connect(unused_client, unused_userdata, unused_flags, return_code):
  """Callback for when a device connects."""
  print('on_connect', mqtt.connack_string(return_code))


def on_disconnect(unused_client, unused_userdata, return_code):
  """Paho callback for when a device disconnects."""
  print('on_disconnect', error_str(return_code))


def on_publish(unused_client, unused_userdata, unused_mid):
  """Paho callback when a message is sent to the broker."""
  print('on_publish')


class CloudIotMqttLogger(CloudIotLogger):
  """A logger that logs to Google Cloud IoT Core via MQTT."""

  GOOGLE_ROOT_CERT_URL = 'https://pki.google.com/roots.pem'
  GOOGLE_ROOT_CERT_FILENAME = 'google_roots.pem'

  # For publishing via MQTT
  DEFAULT_MQTT_BRIDGE = 'mqtt.googleapis.com'
  MQTT_PORTS = [8883, 443]
  DEFAULT_MQTT_PORT = 8883

  def __init__(self, project_id, registry_id, priv_key_file, device_id,
               signing_alg=DEFAULT_ALG, cloud_region=None,
               bridge_hostname=None, bridge_port=None):
    """Creates a logger that logs to Google Cloud IoT Core via MQTT.

    Args:
      project_id: The cloud project ID this device belongs to.
      registry_id: The Cloud IoT Core device registry this device belongs to.
      priv_key_file: Path to a file containing the device's RSA256 or ES256
                     private key. It must have been added to the IoT Core
                     registry.
      device_id: Device ID string. Must be the same as the device ID registered
                 with Cloud IoT Core.
      signing_alg: The signing algorithm used to make a JWT for authentication
                   with Cloud IoT Core. Must be one of {RSA256, ES256}.
      cloud_region: The cloud region of the Cloud IoT Core device registry.
      bridge_hostname: URL of the MQTT bridge. Defaults to mqtt.googleapis.com.
      bridge_port: Port to connect to on the MQTT bridge. Defaults to 8883.

    Raises:
      ValueError: If signing_alg is not one of {RSA256, ES256}.
    """
    if cloud_region is None:
      cloud_region = self.DEFAULT_CLOUD_REGION

    if bridge_hostname is None:
      bridge_hostname = self.DEFAULT_MQTT_BRIDGE

    if bridge_port is None:
      bridge_port = self.DEFAULT_MQTT_PORT

    super(CloudIotMqttLogger, self).__init__(
        project_id, registry_id, priv_key_file, device_id, signing_alg,
        cloud_region)

    self._bridge_hostname = bridge_hostname
    self._bridge_port = bridge_port

  def _get_google_root_certs(self):
    dir_path = os.path.dirname(os.path.realpath(__file__))
    cert_path = os.path.join(dir_path, self.GOOGLE_ROOT_CERT_FILENAME)

    if not os.path.isfile(cert_path):
      response = requests.get(self.GOOGLE_ROOT_CERT_URL, stream=True)
      if response.status_code == 200:
        with open(cert_path, 'wb') as f:
          for chunk in response:
            f.write(chunk)
      else:
        raise Exception('Failed to download Google root certs!')

    return cert_path

  @property
  def _client_id(self):
    """Returns the unique string that identifies this device.

    Google Cloud IoT Core expects a string of this format.

    Returns:
      The MQTT client ID of this device.
    """
    return 'projects/{}/locations/{}/registries/{}/devices/{}'.format(
        self._project_id, self._cloud_region, self._registry_id,
        self._device_id)

  def _get_topic(self, topic_type):
    assert topic_type in ['events', 'state']
    return '/devices/{}/{}'.format(self._device_id, topic_type)

  @property
  def _event_topic(self):
    return self._get_topic('events')

  @property
  def _state_topic(self):
    return self._get_topic('state')

  def _get_client(self):
    """Creates an MQTT client."""
    client = mqtt.Client(client_id=self._client_id)

    # Google Cloud IoT Core ignores the username field and uses the
    # password field to transmit a JWT to authorize the device.
    client.username_pw_set(username='unused', password=self._create_jwt())

    # Enable SSL/TLS support.
    client.tls_set(ca_certs=self._get_google_root_certs(),
                   tls_version=ssl.PROTOCOL_TLSv1_2)

    # Register message callbacks. https://eclipse.org/paho/clients/python/docs/
    # describes additional callbacks that Paho supports. In this example, the
    # callbacks just print to standard out.
    client.on_connect = on_connect
    client.on_publish = on_publish
    client.on_disconnect = on_disconnect

    # Connect to the Google MQTT bridge.
    client.connect(self._bridge_hostname, self._bridge_port)

    return client

  def log(self, timestamp, values):
    client = None
    try:
      client = self._get_client()

      # Start the network loop.
      client.loop_start()

      measurement = self._get_proto(timestamp, values)

      # Publish to the MQTT topic. qos=1 means at least once delivery.
      # Cloud IoT Core also supports qos=0 for at most once delivery.
      client.publish(self._event_topic, measurement.SerializeToString(), qos=1)
    finally:
      if client is not None:
        client.loop_stop()


class CloudIotHttpLogger(CloudIotLogger):
  """A logger that logs to Google Cloud IoT Core via HTTP."""

  PUBLISH_URL = 'https://cloudiotdevice.googleapis.com/v1'

  def __init__(self, project_id, registry_id, priv_key_file, device_id,
               signing_alg=DEFAULT_ALG, cloud_region=None):
    """Creates a logger that logs to Google Cloud IoT Core via HTTP.

    Args:
      project_id: The cloud project ID this device belongs to.
      registry_id: The Cloud IoT Core device registry this device belongs to.
      priv_key_file: Path to a file containing the device's RSA256 or ES256
                     private key. It must have been added to the IoT Core
                     registry.
      device_id: Device ID string. Must be the same as the device ID registered
                 with Cloud IoT Core.
      signing_alg: The signing algorithm used to make a JWT for authentication
                   with Cloud IoT Core. Must be one of {RSA256, ES256}.
      cloud_region: The cloud region of the Cloud IoT Core device registry.

    Raises:
      ValueError: If signing_alg is not one of {RSA256, ES256}.
    """
    if cloud_region is None:
      cloud_region = self.DEFAULT_CLOUD_REGION

    super(CloudIotHttpLogger, self).__init__(
        project_id, registry_id, priv_key_file, device_id, signing_alg,
        cloud_region)

  def _get_publish_url(self, message_type):
    url_suffixes = {
        'event': 'publishEvent',
        'state': 'setState'
    }
    assert message_type in url_suffixes

    return '{}/projects/{}/locations/{}/registries/{}/devices/{}:{}'.format(
        self.PUBLISH_URL, self._project_id, self._cloud_region,
        self._registry_id, self._device_id, url_suffixes[message_type])

  @property
  def _event_publish_url(self):
    return self._get_publish_url('event')

  @property
  def _state_publish_url(self):
    return self._get_publish_url('state')

  def _pack_request_body(self, payload, message_type):
    msg_bytes = base64.urlsafe_b64encode(payload)

    if message_type == 'event':
      return {'binary_data': msg_bytes.decode('ascii')}
    elif message_type == 'state':
      return {'state': {'binary_data': msg_bytes.decode('ascii')}}
    else:
      raise Exception('Unknown message type: "{}"'.format(message_type))

  def log(self, timestamp, values):
    headers = {
        'authorization': 'Bearer {}'.format(self._create_jwt()),
        'content-type': 'application/json',
        'cache-control': 'no-cache'
    }

    measurement = self._get_proto(timestamp, values)
    body = self._pack_request_body(measurement.SerializeToString(), 'event')

    # TODO(mtraver) check return code
    requests.post(
        self._event_publish_url, data=json.dumps(body), headers=headers)
