import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface SensetifQuery extends DataQuery {
  project: string;
  subsystem: string;
  datapoint: string;

  channel?: string;
}

export const defaultQuery: Partial<SensetifQuery> = {
  project: '',
  subsystem: '',
  datapoint: '',

  channel: 'sensetif-channel-default',
};

export interface SensetifDataSourceOptions extends DataSourceJsonData {
  projects: string[];
}

export interface Project {}

export interface PollDeclaration {
  schedule: string;
  zoneId: string;
  organizationName: string;
  project: string;
  subsystemName: string;
  url: string;
  user: string;
  password: string;
  timestampExpression: string;
  timeToLive: TimeToLive;
  timestampType: TimestampType;
}

export enum TimeToLive {
  a, // ninety-two days
  b, // 370 days
  c, // 5.5 years
}

export enum TimestampType {
  epochMillis,
  epochSeconds,
  iso8601_zoned,
  iso8601_offset,
}
