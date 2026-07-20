export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  DateTime: { input: string; output: string; }
};

export type Measurement = {
  __typename: 'Measurement';
  aqi: Maybe<Scalars['Float']['output']>;
  co2: Maybe<Scalars['Float']['output']>;
  deviceId: Scalars['String']['output'];
  hcho: Maybe<Scalars['Float']['output']>;
  noxIndex: Maybe<Scalars['Float']['output']>;
  pm1: Maybe<Scalars['Float']['output']>;
  pm4: Maybe<Scalars['Float']['output']>;
  pm10: Maybe<Scalars['Float']['output']>;
  pm25: Maybe<Scalars['Float']['output']>;
  rh: Maybe<Scalars['Float']['output']>;
  temp: Maybe<Scalars['Float']['output']>;
  timestamp: Scalars['DateTime']['output'];
  uploadTimestamp: Scalars['DateTime']['output'];
  vocIndex: Maybe<Scalars['Float']['output']>;
};

export type Query = {
  __typename: 'Query';
  latest: Array<Measurement>;
  measurements: Array<Measurement>;
};


export type QueryMeasurementsArgs = {
  endTime?: InputMaybe<Scalars['DateTime']['input']>;
  startTime: Scalars['DateTime']['input'];
};

export type MeasurementFieldsFragment = { __typename: 'Measurement', deviceId: string, timestamp: string, uploadTimestamp: string, temp: number | null, pm1: number | null, pm25: number | null, pm4: number | null, pm10: number | null, aqi: number | null, rh: number | null, co2: number | null, vocIndex: number | null, noxIndex: number | null, hcho: number | null };

export type GetMeasurementsQueryVariables = Exact<{
  startTime: Scalars['DateTime']['input'];
  endTime?: InputMaybe<Scalars['DateTime']['input']>;
}>;


export type GetMeasurementsQuery = { measurements: Array<{ __typename: 'Measurement', deviceId: string, timestamp: string, uploadTimestamp: string, temp: number | null, pm1: number | null, pm25: number | null, pm4: number | null, pm10: number | null, aqi: number | null, rh: number | null, co2: number | null, vocIndex: number | null, noxIndex: number | null, hcho: number | null }> };

export type LatestQueryVariables = Exact<{ [key: string]: never; }>;


export type LatestQuery = { latest: Array<{ __typename: 'Measurement', deviceId: string, timestamp: string, uploadTimestamp: string, temp: number | null, pm1: number | null, pm25: number | null, pm4: number | null, pm10: number | null, aqi: number | null, rh: number | null, co2: number | null, vocIndex: number | null, noxIndex: number | null, hcho: number | null }> };
