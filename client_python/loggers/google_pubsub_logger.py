"""Publish to Google Cloud Pub/Sub"""
from google.cloud import pubsub_v1  # pylint: disable=import-error

from . import base_logger


class PubSubLogger(base_logger.GCPLogger):
  """A logger that publishes to Google Cloud Pub/Sub."""

  def __init__(self, project_id, topic, device_id):
    """Creates a logger that publishes to Google Cloud Pub/Sub.

    Args:
      project_id: The ID of the Google Cloud project the device belongs to.
      topic: The Cloud Pub/Sub topic name.
      device_id: Device ID string.
    """
    super(PubSubLogger, self).__init__(project_id, device_id)

    self._topic = topic

  def log(self, timestamp, values):
    publisher = pubsub_v1.PublisherClient()
    topic_path = publisher.topic_path(self._project_id, self._topic)

    publisher.publish(
        topic_path, self._get_proto(timestamp, values).SerializeToString())
