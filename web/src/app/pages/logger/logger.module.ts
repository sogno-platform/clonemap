import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

import { LoggerRoutingModule } from './logger-routing.module';
import { NgxMatTimepickerModule } from 'ngx-mat-timepicker';
import { NgxChartsModule } from '@swimlane/ngx-charts';
import { NgxPaginationModule } from 'ngx-pagination';
import { MatDatepickerModule } from '@angular/material/datepicker';
import { MatNativeDateModule } from '@angular/material/core/';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatIconModule } from '@angular/material/icon';
import { SharedModule } from '../shared/shared.module';


import { LogComponent } from "./log/log.component";
import { LogseriesComponent } from "./logseries/logseries.component";
import { StatsComponent } from './stats/stats.component';
import { HeatmapComponent } from './heatmap/heatmap.component';
import { TabsComponent } from './tabs/tabs.component';
import { AgentSelectorComponent } from './widgets/agent-selector/agent-selector.component';
import { PeriodSelectorComponent } from './widgets/period-selector/period-selector.component';

const materialModules = [
  MatDatepickerModule,
  MatNativeDateModule,
  MatFormFieldModule,
  MatInputModule,
  MatIconModule,
  MatSelectModule,
];


@NgModule({
  declarations: [
    LogComponent,
    LogseriesComponent,
    StatsComponent,
    HeatmapComponent,
    TabsComponent,
    AgentSelectorComponent,
    PeriodSelectorComponent,   
  ],
  imports: [
    CommonModule,
    LoggerRoutingModule,
    ...materialModules,
    NgxMatTimepickerModule,
    NgxChartsModule,
    NgxPaginationModule,
    FormsModule,
    ReactiveFormsModule,
    SharedModule,
  ],
  exports: [
    LogComponent,
    LogseriesComponent,
  ]
})
export class LoggerModule { 
}
