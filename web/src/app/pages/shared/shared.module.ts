import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SidebarComponent } from './sidebar/sidebar.component';
import { TopBarComponent } from './top-bar/top-bar.component';
import { HttpClientModule} from '@angular/common/http';
import { RouterModule } from '@angular/router';


@NgModule({
  declarations: [
    SidebarComponent,
    TopBarComponent,
  ],
  imports: [
    CommonModule,
    HttpClientModule,
    RouterModule,
  ],
  exports: [
    SidebarComponent,
    TopBarComponent,
  ]
})
export class SharedModule { }
