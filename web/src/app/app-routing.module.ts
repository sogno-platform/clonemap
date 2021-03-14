import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { OverviewComponent} from './pages/overview/overview.component';
import { AMSComponent} from './pages/ams/ams.component';
import { DFComponent} from './pages/df/df.component';
import { LoggerComponent} from './pages/logger/logger.component';


const routes: Routes = [
  { path: '', redirectTo: '/overview', pathMatch: 'full'},
  { path: 'overview', component: OverviewComponent },
  { path: 'ams', component: AMSComponent},
  { path: 'ams/:masId', component: AMSComponent},
  { path: 'df', component: DFComponent},
  { path: 'logger', component: LoggerComponent},
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
