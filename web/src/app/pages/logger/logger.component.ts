import { Component, OnInit } from '@angular/core';
import { LoggerService} from 'src/app/services/logger.service'
import { MasService } from 'src/app/services/mas.service';

@Component({
  selector: 'app-logger',
  templateUrl: './logger.component.html',
  styleUrls: ['./logger.component.css']
})
export class LoggerComponent implements OnInit {

  alive: boolean = false;
  selectedMASId = -1;
  MASs;

  constructor(
    private loggerService: LoggerService,
    private masService: MasService
    ) { }

  ngOnInit(): void {
    this.loggerService.getAlive().subscribe( (res: any) => {
      if (res.logger) {
        this.alive = res.logger ;
      }
    }, 
    error => {
      console.log(error);
  });
        // update the sidebar
        this.masService.getMAS().subscribe((MASs: any) => {
          if (MASs === null) {
              console.log(status);
              this.MASs = [];
          } else {
              this.MASs = MASs
          } 
          }, 
          err => {
              console.log(err)  
          }
      );

  }

}
