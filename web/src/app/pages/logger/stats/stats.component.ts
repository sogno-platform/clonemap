import { Component, OnInit } from '@angular/core';
import { StatsInfo } from 'src/app/models/stats.mdoel'
import { Router, Event, NavigationEnd } from '@angular/router';
import { DefaultAMSService } from 'src/app/openapi-services/ams';
import { DefaultLoggerService } from 'src/app/openapi-services/logger';

@Component({
  selector: 'app-stats',
  templateUrl: './stats.component.html',
  styleUrls: ['./stats.component.scss']
})
export class StatsComponent implements OnInit {

    selectedMASID: number = -1;
    agentID: number[] = [];

    agentStats: number = -1;
    behaviorTypes: string[] = ["mqtt", "protocol", "period", "custom"]
    selectedBehType: string = "mqtt";
    statsInfo: StatsInfo = null;
    q: number = 1;


    constructor(
        private loggerService: DefaultLoggerService,
        private amsService: DefaultAMSService,
        private router: Router,
    ) { 
        this.router.events.subscribe((event: Event) => {
            if (event instanceof NavigationEnd) {
                this.selectedMASID = Number(this.router.url.split("/")[2]);
            }
        });
    }

    ngOnInit(): void {
        this.amsService.getOneMAS(this.selectedMASID).subscribe((res: any) => {
            if (res.agents.counter !== 0) {
                this.agentID = res.agents.instances.map(item => item.id);
            }
        });
    }

    onClickSearchButton(
        selectedStartDate,
        selectedEndDate,
        selectedStartTime,
        selectedEndTime
    ){ 
        const startDate: string = this.convertDate(selectedStartDate);
        const endDate:string =  this.convertDate(selectedEndDate);
        const searchStartTime = startDate + selectedStartTime.replace(":", "") + "00";
        const searchEndTime = endDate + selectedEndTime.replace(":", "") + "59";
        this.drawStatistics(searchStartTime, searchEndTime);   
    }

    updateSelectedBehType(beh: string) {
        this.selectedBehType = beh;
    }

    //display Statistics infomation
    drawStatistics(searchStartTime, searchEndTime) {
        if (this.agentStats !== -1) {
            const methods: string[] = ["max", "min", "count", "average"];
            this.loggerService.getStats(this.selectedMASID, this.agentStats,
            this.selectedBehType, searchStartTime, searchEndTime).subscribe( (res: any) => {
                this.statsInfo = res;
                if (this.statsInfo.list === null) {
                    this.statsInfo.list = [];
                }
            })
        }
    }

    updateSelectedAgent(agentID: number) {
        this.agentStats = agentID;
    }

    convertDate(date: Date): string {
        let res: string = date.getFullYear().toString();
        res += (date.getMonth() + 1) < 10 ? "0" + (date.getMonth() + 1).toString() : (date.getMonth() + 1).toString()
        res += date.getDate()  < 10 ? "0" + date.getDate().toString() : date.getDate().toString()
        return res;
    }

    // convertSecond converts second to minute, hour, day when necessary
    convertSecond(origin: number): string {
        const hour: number = Math.floor(origin / (60 * 60));
        const minute: number = Math.floor(origin % (60 * 60) /60);
        const second: number = Math.floor(origin % 60);
        let res: string = "";
        res += ' ' +  hour.toString() + "H"
        res += ' ' + minute.toString() + "M"
        res += ' ' + second.toString() + "S"
        return res;
    }

}
