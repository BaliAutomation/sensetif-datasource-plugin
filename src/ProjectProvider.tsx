import React, { PureComponent } from 'react';
import { getBackendSrv } from '@grafana/runtime';
import { Project } from './types';

export interface ProjectAPI {
  loadProjects: () => void;
}

export interface LoadingStates {
  loadProjects: boolean;
}

export interface Props {
  children: (api: ProjectAPI, states: LoadingStates, projects: Project[]) => JSX.Element;
}

export interface State {
  projects: Project[];
  loadingStates: LoadingStates;
}

export class ProjectProvider extends PureComponent<Props, State> {
  state: State = {
    projects: [] as Project[],
    loadingStates: {
      loadProjects: true,
    },
  };

  loadProject = async () => {
    this.setState({
      loadingStates: { ...this.state.loadingStates, loadProjects: true },
    });
    const projects = await getBackendSrv().get('/api/plugins/sensetif-datasource/resources/projects');
    this.setState({
      projects,
      loadingStates: { ...this.state.loadingStates, loadProjects: Object.keys(projects).length === 0 },
    });
  };

  render() {
    const { children } = this.props;
    const { loadingStates, projects } = this.state;
    const api: ProjectAPI = {
      loadProjects: this.loadProject,
    };
    return <>{children(api, loadingStates, projects)}</>;
  }
}

export default ProjectProvider;
