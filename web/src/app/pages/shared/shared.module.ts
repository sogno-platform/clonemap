import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SidebarComponent } from './sidebar/sidebar.component';
import { TopBarComponent } from './top-bar/top-bar.component';
import { HttpClientModule} from '@angular/common/http';
import { RouterModule } from '@angular/router';

import { MatDatepickerModule } from '@angular/material/datepicker';
import { MatNativeDateModule } from '@angular/material/core/';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { MatPaginatorModule } from '@angular/material/paginator';




const materialModules = [
  MatDatepickerModule,
  MatNativeDateModule,
  MatFormFieldModule,
  MatInputModule,
  MatIconModule,
  MatSelectModule,
  MatTableModule,
  MatPaginatorModule,
];


@NgModule({
  declarations: [
    SidebarComponent,
    TopBarComponent,
  ],
  imports: [
    CommonModule,
    HttpClientModule,
    RouterModule,
    ...materialModules,
  ],
  exports: [
    SidebarComponent,
    TopBarComponent,
    ...materialModules,
  ]
})        
export class SharedModule { }
