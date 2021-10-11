import { NgModule } from "@angular/core";
import { Routes, RouterModule } from '@angular/router';

import { LogComponent } from "./log/log.component";
import { LogseriesComponent } from "./logseries/logseries.component";
import { StatsComponent } from './stats/stats.component';
import { HeatmapComponent } from './heatmap/heatmap.component';
import { AgentSelectorComponent } from './widgets/agent-selector/agent-selector.component';

const routes: Routes = [
  { path: "logger/:masid", component: LogComponent},
  { path: "log/:masid", component: LogComponent },
  { path: "logseries/:masid", component: LogseriesComponent },
  { path: "stats/:masid", component: StatsComponent },
  { path: "heatmap/:masid", component: HeatmapComponent },
]

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})

export class LoggerRoutingModule{ }