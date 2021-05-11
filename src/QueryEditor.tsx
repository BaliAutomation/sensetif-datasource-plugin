import defaults from 'lodash/defaults';

import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from './datasource';
import { defaultQuery, SensetifDataSourceOptions, SensetifQuery } from './types';

const { FormField } = LegacyForms;

type Props = QueryEditorProps<DataSource, SensetifQuery, SensetifDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  onQueryProjectChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, project: event.target.value });
  };
  onQuerySubsystemChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, subsystem: event.target.value });
  };
  onQueryDatapointChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, datapoint: event.target.value });
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { project, subsystem, datapoint } = query;

    return (
      <div className="gf-form">
        <FormField
          labelWidth={8}
          value={project}
          onChange={this.onQueryProjectChange}
          label="Project"
          tooltip="The project to be queried"
        />
        <FormField
          labelWidth={8}
          value={subsystem}
          onChange={this.onQuerySubsystemChange}
          label="Subsystem"
          tooltip="The Subsystem within the project to be queried."
        />
        <FormField
          labelWidth={8}
          value={datapoint}
          onChange={this.onQueryDatapointChange}
          label="Datapoint"
          tooltip="The Datapoint in the Subsystem."
        />
      </div>
    );
  }
}
