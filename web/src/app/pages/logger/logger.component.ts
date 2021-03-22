import { Component, OnInit } from '@angular/core';
import { LoggerService} from 'src/app/services/logger.service'
import { MasService } from 'src/app/services/mas.service';
import { ActivatedRoute, Params} from '@angular/router'

@Component({
  selector: 'app-logger',
  templateUrl: './logger.component.html',
  styleUrls: ['./logger.component.css']
})
export class LoggerComponent implements OnInit {

    alive: boolean = false;
    selectedMASId: number = -1;
    MASs = null;
    searched_results;

    constructor(
        private loggerService: LoggerService,
        private masService: MasService,
        private route: ActivatedRoute
        ) { }

    ngOnInit(): void {

        // check whether the logger is alive
        this.loggerService.getAlive().subscribe( (res:any) => {
            this.alive = res.logger;
        });

        // update the sidebar
        this.masService.getMAS().subscribe((MASs: any) => {
            this.MASs = MASs;
            },
            err => {
                console.log(err)  
        });
        
        // get the selectedMASid from the current route
        this.route.params.subscribe((params: Params) => {
                this.selectedMASId = params.masId;             
            });
    }

        onSearchLogs(agentid:string, topic:string, num: string) {
            this.loggerService.getNLatestLogger(this.selectedMASId.toString(), agentid, topic, num)
                .subscribe( res => {
                    this.searched_results = res;
                },
                err => console.log(err)
                )
            }
}
