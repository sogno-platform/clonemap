import { Component, OnInit } from '@angular/core';
import { LoggerService} from 'src/app/services/logger.service'

@Component({
  selector: 'app-logger',
  templateUrl: './logger.component.html',
  styleUrls: ['./logger.component.css']
})
export class LoggerComponent implements OnInit {

  status: string = "I am not alive";

  constructor(private loggerService: LoggerService) { }

  ngOnInit(): void {
    this.loggerService.getAlive().subscribe( res => {
      this.status =  res.toString();
  }, error => {
      console.log(error);
  });
  }

}
