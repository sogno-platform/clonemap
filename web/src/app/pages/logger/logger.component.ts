import { Component, OnInit} from '@angular/core';
import { LoggerService} from 'src/app/services/logger.service'
import { MasService } from 'src/app/services/mas.service';
import { ActivatedRoute, Params} from '@angular/router';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { LogMessage, LogSeries, pointSeries } from 'src/app/models/log.model';
import { forkJoin, Observable} from 'rxjs';
import { FormControl, FormGroup } from '@angular/forms';



@Component({
  selector: 'app-logger',
  templateUrl: './logger.component.html',
  styleUrls: ['./logger.component.css']
})
export class LoggerComponent implements OnInit {

    alive: boolean = true;
    selectedMASID: number = -1;
    MASID: number[] = [];
    currState: string = "log";

    agentID: number[];
    selectedID: number[] = [];
    notSelectedID: number[] = [];
    isAgentSelected: boolean[] = [];

    // parameters and variables for drawing logs
    searchStartTime: string = "20210301000000";
    searchEndTime: string = "20210331000000"
    isTopicSelected: boolean[] = [true, true, true, true, true, true];
    topics: string[] = ["error", "debug", "msg", "status", "app", "beh" ];
    width: number = 1500;
    height: number = 2000;
    boxWidth: number = 100;
    boxHeight: number = 50;
    logBoxWidth: number = 50;
    logBoxHeight: number = 25;
    interval: number;
    timeline: any = [];
    agentBox: any = [];
    logBoxes: any = [];
    communications: any = [];
    texts = [];  
    popoverContent: string = "This is the content of the popover";
    logs: LogMessage[] = [];
    range = new FormGroup( {
        start: new FormControl(),
        end: new FormControl()
    });
    selectedStartDate: Date = new Date();
    selectedEndDate: Date = new Date();
    selectedStartTime: string = "00:00";
    selectedEndTime: string = "23:59";

    // parameters and variables for drawing log series
    selectedName: string = "";
    names: string[] = [];
    logSeries: LogSeries[] = [];
    bubbleData: pointSeries[] = [];
    view: any[]= [1600, 600];
    showXAxis: boolean = true;
    showYAxis: boolean = true;
    gradient: boolean = false;
    showLegend: boolean = false;
    showXAxisLabel: boolean = true;
    xAxisLabel: string = 'Time';
    showYAxisLabel: boolean = true;
    yAxisLabel: string = 'Value';
    maxRadius: number = 10;
    minRadius: number = 10;
    datesSeries: Date[] = [];
    scaledDatesSeries: number[];
    xAxisTicks: number[] = [];
    mapAxisDate: Map<number, string> = new Map<number, string>();
    xAxisTickFormatting = (val:number) => {
      return this.mapAxisDate.get(val);
    }
    colorScheme = {
      domain: ['#9696F3'
/*       'rgb(230, 109, 109)', 
      'rgb(216, 216, 110)', 
      'rgb(126, 235, 149)', 
      'rgb(56, 202, 221)'  */]
    };

    constructor(
        private loggerService: LoggerService,
        private masService: MasService,
        private route: ActivatedRoute,
        private modalService: NgbModal,
        ) {}

    ngOnInit(): void {

        // check whether the logger is alive
        this.loggerService.getAlive().subscribe( (res:any) => {
            this.alive = res.logger;
        });

        // update the sidebar
        this.masService.getMAS().subscribe((MASs: any) => {
            if (MASs !== null) {
                this.MASID = MASs.map(MAS => MAS.id);
            } 
        }, err => {
            console.log(err);
        });

        // get the selectedMASid from the current route
        this.route.params.subscribe((params: Params) => {
            this.selectedMASID = params.masid;  
            this.masService.getMASById(params.masid).subscribe((res: any) => {
                if (res.agents.counter !== 0) {
                    this.agentID = res.agents.instances.map(item => item.id);
                    for (let i = 0; i < res.agents.counter; i++) {
                        this.isAgentSelected.push(false);
                    }
                    this.updateSelectedID();
                }
            });
        });   
    } 
    


    /********************************** common functions ************************************/
    onDeleteID(i : number) {
        this.isAgentSelected[i] = !this.isAgentSelected[i];
        this.updateSelectedID();
        this.updateNames();
        if (this.currState === "log") {
            this.drawLogs();
        } else {
            this.drawSeries();
        }
    }

    openLg(content) {
        this.modalService.open(content, { size: 'lg', centered: true });
    }

    onAddID(i: number) {
        if (this.selectedID.length < 10) {
            this.isAgentSelected[i] = !this.isAgentSelected[i];
            this.updateSelectedID();
            this.updateNames();
        }
    }

    onConfirm() {
        this.modalService.dismissAll();
        if (this.currState === "log") {
            this.drawLogs();
        } else {
            this.drawSeries();
        }
    }

    updateSelectedID() {
        this.selectedID = [];
        this.notSelectedID = [];
        for (let i = 0; i < this.agentID.length; i++) {
            if (this.isAgentSelected[i]) {
                this.selectedID.push(i);
            } else {
                this.notSelectedID.push(i);
            }
        }
    }

    generateScaledDates(dates: Date[]) :number[]{       
        // find the date differences
        let datesInterval: number[] = [0];
        for (let i = 1; i < dates.length; i++) {
            datesInterval.push(Math.round(dates[i-1].getTime()/1000) - Math.round(dates[i].getTime()/1000));
        }

        // find the maximum and minimum interval
        let minDiff : number = Number.MAX_SAFE_INTEGER;
        let maxDiff : number = 0

        for (let i = 0; i < datesInterval.length; i++) {
            if (datesInterval[i] > maxDiff) {
                maxDiff = datesInterval[i];
            }
            if (datesInterval[i] !== 0 && datesInterval[i] < minDiff) {
                minDiff = datesInterval[i];
            }
        }

        // generate  scaledDates
        let scaledDates: number[] = [];
        let curr:number = 0;
        if (maxDiff !== minDiff) {
            for (let i = 0; i < datesInterval.length; i++) {
                if (datesInterval[i] !== 0) {
                    curr = curr +  Math.round(100 * ((5 - 1)  * (datesInterval[i] - minDiff)/(maxDiff - minDiff) + 1)) / 100;
                }
                scaledDates.push(curr); 
            }
        } else {
            for (let i = 0; i < datesInterval.length; i++) {
                if (datesInterval[i] !== 0) {
                    curr = curr + 1
                }
                scaledDates.push(curr);
            }
        }
        return scaledDates;
    }

    onClickSearchButton() {
        const startDate: string = this.convertDate(this.selectedStartDate);
        const endDate:string =  this.convertDate(this.selectedEndDate);
        this.searchStartTime = startDate + this.selectedStartTime.replace(":", "") + "00";
        this.searchEndTime = endDate + this.selectedEndTime.replace(":", "") + "59";
        if (this.currState==="log") {
            this.drawLogs();
        } else {
            this.drawSeries();
        }
    }



    convertDate(date: Date): string {
        let res: string = date.getFullYear().toString();
        res += (date.getMonth() + 1) < 10 ? "0" + (date.getMonth() + 1).toString() : (date.getMonth() + 1).toString()
        res += date.getDate()  < 10 ? "0" + date.getDate().toString() : date.getDate().toString()
        return res;
    }

    /********************************* functions for drawing logs  ************************************/
    onClickLog(){
        this.currState = "log";
    }

    onToggleTopic(i: number) {
        this.isTopicSelected[i] = ! this.isTopicSelected[i];
        this.updateScaledDates();
    }

    drawLogs() {
        this.logs = [];      
        this.multiLogs().subscribe( logss => {
            for (let logs of logss) {
                if (logs !== null) {
                    for (let log of logs) {
                        this.logs.push(log);
                    }
                }
            }
            this.logs.sort((a, b) => {
                let date1 = new Date(a.timestamp);
                let date2 = new Date(b.timestamp);
                return date2.getTime() - date1.getTime();
            })
            this.drawAllElements(this.logs);
         })
    }

    multiLogs(): Observable<any[]> {
        let res = [];
        for (let id of this.selectedID) {
            for (let topic of this.topics) {
                res.push(this.loggerService.getLogsInRange(this.selectedMASID.toString(),
                id.toString(), topic, this.searchStartTime, this.searchEndTime));
            }
        }
        return forkJoin(res);
    } 

    drawAgentBox() {
        this.agentBox = [];
        this.texts = [];
        this.interval = 1 / (1 + this.selectedID.length) * this.width;

        for (let i=0; i < this.selectedID.length; i++) {
            const X: number = (i+1) * this.interval;
            // plot the agent box
            this.agentBox.push({x: X - this.boxWidth / 2, y: 200 - this.boxHeight});
            this.texts.push({
                x: X - this.boxWidth * 5 / 12, 
                y: 200 - this.boxHeight / 3,
                textID: this.selectedID[i],
            })
        }
    }

    drawTimeline() {
        this.timeline = []
        for (let i = 0; i < this.selectedID.length; i++) {
            let X = (i+1) * this.interval;
            this.timeline.push({x1: X, y1:200, x2: X, y2: this.height })  
        }
    }

    drawScaledDates(scaledDates: number[]) {
        this.logBoxes = [];
        this.communications = [];
        for (let i = 0; i < scaledDates.length; i++) {
            let currMsg = this.logs[i];
            let idx = this.selectedID.indexOf(currMsg.agentid) + 1;        
            this.logBoxes.push({
                x: this.interval *idx - this.logBoxWidth / 2, 
                y: 400 + this.logBoxHeight * scaledDates[i] * 1.1,
                topic: currMsg.topic,
                timestamp: currMsg.timestamp,
                msg: currMsg.msg,
                data: currMsg.data,
                hidden: !this.isTopicSelected[this.topics.indexOf(currMsg.topic)],
                
            });
            if (currMsg.topic === "msg" && currMsg.msg ==="ACL send"){
                const data = this.logs[i].data.split(";");
                const sender = Number(data[0].charAt(data[0].length - 1));
                const senderIdx = this.selectedID.indexOf(sender) + 1;
                const receiver = Number(data[1].charAt(data[1].length - 1));
                const receiverIdx = this.selectedID.indexOf(receiver) + 1;
                const direction = (senderIdx < receiverIdx) ? 1 : -1;
                if (this.selectedID.includes(receiver) && this.selectedID.includes(sender)) {
                    this.communications.push({
                        x1: this.interval * senderIdx + direction * this.logBoxWidth / 2,
                        y1: 400 + this.logBoxHeight * scaledDates[i] * 1.1 + this.logBoxHeight / 2,
                        x2: this.interval * receiverIdx - direction * this.logBoxWidth / 2,
                        y2: 400 +   this.logBoxHeight * scaledDates[i] * 1.1 + this.logBoxHeight / 2,
                        hidden: !this.isTopicSelected[this.topics.indexOf("msg")],
                    })
                }
            }
        }
    }

    updateScaledDates() {
        for (let i = 0; i < this.logBoxes.length; i++) {
            let idx = this.topics.indexOf(this.logBoxes[i].topic);
            this.logBoxes[i].hidden = !this.isTopicSelected[idx];
        }
        
        for (let i = 0; i < this.communications.length; i++) {
            let idx = this.topics.indexOf("msg")
            this.communications[i].hidden = !this.isTopicSelected[idx];
        }
    }
                
    drawAllElements(msgs: LogMessage[]) {
        this.drawAgentBox();
        let dates: Date[] = []
        for (let i = 0; i < this.logs.length; i++) {
            let date = new Date(msgs[i].timestamp)
            dates.push(date)
        }
        let scaledDates: number[] = this.generateScaledDates(dates);
        this.height = 800 + this.logBoxHeight * scaledDates[scaledDates.length-1];
        this.drawScaledDates(scaledDates);
        this.drawTimeline();
    }

    onChangePopoverContent(i) {
        this.popoverContent = this.logs[i].msg;
    }


    /********************************  functions for drawing log series   ********************************/ 
    
    onClickLogSeries(){
        this.currState = "logSeries";
        this.drawSeries();
    }
    
    updateSelectedName(name) {
        this.selectedName = name;
    }

    updateNames() {
        this.names = [];
        this.multiNames().subscribe( namess => {
            for (let names of namess) {
                if ( names!== null) {
                    for (let name of names) {
                        if (!this.names.includes(name)) {
                            this.names.push(name)
                        }
                    }
                }
            }
        })
    }

    multiNames(): Observable<any> {
        let res = [];
        for (let id of this.selectedID) {
            res.push(this.loggerService.getLogSeriesNames(this.selectedMASID.toString(),
            id.toString()));
        }
        return forkJoin(res);      
    }

    multiSeries(): Observable<any> {
        let res = [];
        for (let id of this.selectedID) {
            res.push(this.loggerService.getLogSeriesByName(this.selectedMASID.toString(),
            id.toString(), this.selectedName, this.searchStartTime, this.searchEndTime));
        }
        return forkJoin(res);
    }

    drawSeries() {   
        this.logSeries = [];
        this.multiSeries().subscribe( logss => {
            for (let logs of logss) {
                if (logs !== null) {
                    for (let log of logs) {
                        this.logSeries.push(log);
                    }
                }
            }
            this.logSeries.sort((a, b) => {
                let date1 = new Date(a.timestamp);
                let date2 = new Date(b.timestamp);
                return date2.getTime() - date1.getTime();
            });
            this.drawSeriesHelper();
        });
    }

    drawSeriesHelper() {
        this.datesSeries = [];
        this.bubbleData = [];
        for (let i = 0; i < this.logSeries.length; i++) {
            let date = new Date(this.logSeries[i].timestamp)
            this.datesSeries.push(date)
        }
        this.scaledDatesSeries = this.generateScaledDates(this.datesSeries); 
        console.log(this.scaledDatesSeries);
        const maxDate : number = this.scaledDatesSeries[this.scaledDatesSeries.length - 1]
        for (let i = 0; i < this.logSeries.length; i++) {
            const x: number = maxDate - this.scaledDatesSeries[i];
            const point = {
                name: new Date(this.logSeries[i].timestamp).toLocaleString('de-DE',{ hour12: false }),
                x: x,
                y: this.logSeries[i].value,
                r: 10
            } 
            this.xAxisTicks.push(x)
            console.log(typeof x)
            this.mapAxisDate.set(x, new Date(this.logSeries[i].timestamp).toLocaleString('de-DE',{ hour12: false }) )
            this.bubbleData.push({
                name: this.logSeries[i].name,
                series: [point]
            })
        }
        console.log(this.mapAxisDate);
    }

    onSelect(data): void {
        console.log('Item clicked', JSON.parse(JSON.stringify(data)));
    }

    onActivate(data): void {
        console.log('Activate', JSON.parse(JSON.stringify(data)));
    }

    onDeactivate(data): void {
        console.log('Deactivate', JSON.parse(JSON.stringify(data)));
    }

    /********************************* functions for drawing logs  ************************************/
    onClickStatistics() {
        this.currState = "staticstics";
    }

}
