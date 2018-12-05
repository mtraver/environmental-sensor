# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: measurement.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import descriptor_pb2 as google_dot_protobuf_dot_descriptor__pb2
from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='measurement.proto',
  package='measurement',
  syntax='proto3',
  serialized_options=_b('Z\013measurement'),
  serialized_pb=_b('\n\x11measurement.proto\x12\x0bmeasurement\x1a google/protobuf/descriptor.proto\x1a\x1fgoogle/protobuf/timestamp.proto\"\x7f\n\x0bMeasurement\x12\x33\n\tdevice_id\x18\x01 \x01(\tB \x82\xb5\x18\x1c^[a-z][a-z0-9+.%~_-]{2,254}$\x12-\n\ttimestamp\x18\x02 \x01(\x0b\x32\x1a.google.protobuf.Timestamp\x12\x0c\n\x04temp\x18\x03 \x01(\x02:.\n\x05regex\x12\x1d.google.protobuf.FieldOptions\x18\xd0\x86\x03 \x01(\tB\rZ\x0bmeasurementb\x06proto3')
  ,
  dependencies=[google_dot_protobuf_dot_descriptor__pb2.DESCRIPTOR,google_dot_protobuf_dot_timestamp__pb2.DESCRIPTOR,])


REGEX_FIELD_NUMBER = 50000
regex = _descriptor.FieldDescriptor(
  name='regex', full_name='measurement.regex', index=0,
  number=50000, type=9, cpp_type=9, label=1,
  has_default_value=False, default_value=_b("").decode('utf-8'),
  message_type=None, enum_type=None, containing_type=None,
  is_extension=True, extension_scope=None,
  serialized_options=None, file=DESCRIPTOR)


_MEASUREMENT = _descriptor.Descriptor(
  name='Measurement',
  full_name='measurement.Measurement',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='device_id', full_name='measurement.Measurement.device_id', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\202\265\030\034^[a-z][a-z0-9+.%~_-]{2,254}$'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='timestamp', full_name='measurement.Measurement.timestamp', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='temp', full_name='measurement.Measurement.temp', index=2,
      number=3, type=2, cpp_type=6, label=1,
      has_default_value=False, default_value=float(0),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=101,
  serialized_end=228,
)

_MEASUREMENT.fields_by_name['timestamp'].message_type = google_dot_protobuf_dot_timestamp__pb2._TIMESTAMP
DESCRIPTOR.message_types_by_name['Measurement'] = _MEASUREMENT
DESCRIPTOR.extensions_by_name['regex'] = regex
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

Measurement = _reflection.GeneratedProtocolMessageType('Measurement', (_message.Message,), dict(
  DESCRIPTOR = _MEASUREMENT,
  __module__ = 'measurement_pb2'
  # @@protoc_insertion_point(class_scope:measurement.Measurement)
  ))
_sym_db.RegisterMessage(Measurement)

google_dot_protobuf_dot_descriptor__pb2.FieldOptions.RegisterExtension(regex)

DESCRIPTOR._options = None
_MEASUREMENT.fields_by_name['device_id']._options = None
# @@protoc_insertion_point(module_scope)
