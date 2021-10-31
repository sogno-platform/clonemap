import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { HttpClientModule, HTTP_INTERCEPTORS} from '@angular/common/http';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { OverviewComponent } from './pages/overview/overview.component';
import { AMSComponent } from './pages/ams/ams.component';
import { DFComponent } from './pages/df/df.component';
import { LoggerModule } from './pages/logger/logger.module';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { NgxMatTimepickerModule } from 'ngx-mat-timepicker';
import { NgxChartsModule } from '@swimlane/ngx-charts';
import { NgxPaginationModule } from 'ngx-pagination';
import { SharedModule } from './pages/shared/shared.module';

import { AgencyApiModule } from './openapi-services/agency';
import { AMSApiModule } from './openapi-services/ams';
import { DFApiModule } from './openapi-services/df';
import { LoggerApiModule } from './openapi-services/logger';


const openAPIModules = [
  AgencyApiModule,
  AMSApiModule,
  DFApiModule,
  LoggerApiModule,
]

@NgModule({
  declarations: [
    AppComponent,
    OverviewComponent,
    AMSComponent,
    DFComponent,
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    NgbModule,
    HttpClientModule,
    BrowserAnimationsModule,
    FormsModule,
    ReactiveFormsModule,
    NgxMatTimepickerModule,
    NgxChartsModule,
    NgxPaginationModule,
    LoggerModule,
    SharedModule,
    ...openAPIModules,
  ],
  exports: [
    SharedModule,
    openAPIModules,
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
