import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { HttpClientModule, HTTP_INTERCEPTORS} from '@angular/common/http';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { OverviewComponent } from './pages/overview/overview.component';
import { AMSComponent } from './pages/ams/ams.component';
import { LoggerComponent } from './pages/logger/logger.component';
import { DFComponent } from './pages/df/df.component';
import { TopBarComponent } from './pages/top-bar/top-bar.component';

@NgModule({
  declarations: [
    AppComponent,
    OverviewComponent,
    AMSComponent,
    LoggerComponent,
    DFComponent,
    TopBarComponent,
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    NgbModule,
    HttpClientModule
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
