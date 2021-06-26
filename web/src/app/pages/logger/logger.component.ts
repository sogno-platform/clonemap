import { Component, OnInit, ViewChild} from '@angular/core';
import { LoggerService} from 'src/app/services/logger.service'
import { MasService } from 'src/app/services/mas.service';
import { ActivatedRoute, Params} from '@angular/router';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { LogMessage, LogSeries, pointSeries } from 'src/app/models/log.model';
import { StatsInfo } from 'src/app/models/stats.mdoel'
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
    tabs: string[] = ["log", "logSeries", "statistics", "heatmap"]
    behaviorTypes: string[] = ["mqtt", "protocol", "period", "custom"]

    // parameters and variables for drawing logs
    searchStartTime: string = "20210301000000";
    searchEndTime: string = "20211231000000"
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
    popoverContent: string[] = ["This is the content of the popover"];
    logs: LogMessage[] = [];
    range = new FormGroup( {
        start: new FormControl(),
        end: new FormControl()
    });
    selectedStartDate: Date = new Date();
    selectedEndDate: Date = new Date();
    selectedStartTime: string = "00:00";
    selectedEndTime: string = "23:59";

    // parameters and variables for ngx-charts
    selectedName: string = "";
    names: string[] = [];
    logSeries: LogSeries[] = [];
    bubbleData: pointSeries[] = [];
    maxRadius: number = 10;
    minRadius: number = 10;
    datesSeries: Date[] = [];
    scaledDatesSeries: number[];
    xAxisTicks: number[] = [];
    mapAxisDate: Map<number, string> = new Map<number, string>();
    xAxisTickFormatting = (val:number) => {
      return this.mapAxisDate.get(val);
    }

    // exclusive for log series
    viewBubble: any[]= [1600, 600];
    colorScheme = {
        domain: [
            "rgb(230, 109, 109)",
            "rgb(230, 178, 109)",
            "rgb(222, 230, 109)",
            "rgb(147, 230, 109)",
            "rgb(109, 230, 210)",
            "rgb(109, 151, 230)",
            "rgb(131, 109, 230)",
            "rgb(196, 109, 230)",
            "rgb(230, 109, 174)",
            "rgb(12, 6, 6)" ]
      };
  
    animations: boolean = true;
    
    // exclusive for statistics
    agentStats: number = -1;
    selectedBehType: string = "mqtt";
    statsInfo: StatsInfo = null;
    q: number = 1;

    // exclusive for heatmap
    gridColors: string[] = []
    grids: any[] = [];
    gridWidth: number = 4;
    popoverFrequency: any = {};
    autoPartition: boolean = false;
    partitionNum: number = 5;
    colorPartitionEle: string[] = [];
    colorLegendTexts: string[] = [];
    legendWidth = 30;
    

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
                for (let MAS of MASs) {
                    if (MAS.status.code != 5) {
                        this.MASID.push(MAS.id)
                    }
                }
            } 
        }, err => {
            console.log(err);
        });


        // get the selectedMASid from the current route
        this.route.params.subscribe((params: Params) => {
            this.selectedMASID = params.masid;  
            this.masService.getMASById(params.masid).subscribe((res: any) => {
                if (res.agents.counter !== 0) {
                    this.gridWidth = 600 / res.agents.counter;
                    console.log(this.gridWidth);
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

    //tabs: string[] = ["log", "logSeries", "statistics", "heatmap"]
    onConfirm() {
        this.modalService.dismissAll();
        switch (this.currState) {
            case this.tabs[0]:
                this.drawLogs();
                break;
            case this.tabs[1]:
                this.drawSeries();
                break;
            case this.tabs[2]:
                this.drawStatistics;
                break;
            default:
                
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
        } else if (this.currState==="logSeries") {
            this.drawSeries();
        } else if (this.currState==="statistics")  {
            this.drawStatistics();
        } else {
            this.drawHeatmap();
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
            this.drawAllElements();
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
        console.log(this.logs);
        console.log(scaledDates);
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
                const sender = Number(data[0].split(" ")[1]);
                const senderIdx = this.selectedID.indexOf(sender) + 1;
                const receiver = Number(data[1].split(" ")[2]);
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
                
    drawAllElements() {
        this.drawAgentBox();
        let dates: Date[] = []
        for (let i = 0; i < this.logs.length; i++) {
            let date = new Date(this.logs[i].timestamp)
            dates.push(date)
        }
        
        let scaledDates: number[] = this.generateScaledDates(dates);
        this.height = 800 + this.logBoxHeight * scaledDates[scaledDates.length-1];
        this.drawScaledDates(scaledDates);
        this.drawTimeline();
    }



    onChangePopoverContent(i) {
        if ("data" in this.logs[i]) {
            if (this.logs[i].msg === "ACL send" || this.logs[i].msg === "ACL receive") {
                this.popoverContent = this.logs[i].data.split(";");
                this.popoverContent[2] = this.popoverContent[2].split(".")[0];
                console.log(this.logs[i].data.split("; "))   
            } else {
                this.popoverContent = [this.logs[i].timestamp.toString().split(".")[0], this.logs[i].msg]
                let data: string[] = this.logs[i].data.split(";");
                data[0] = data[0].split(".")[0];
                data[1] = data[1].split(".")[0];
                this.popoverContent = [...this.popoverContent, ...data]

            }
        } else {
            this.popoverContent = [this.logs[i].timestamp.toString().split(".")[0], ...this.logs[i].msg.split(";")];
        }
    }


    /********************************  functions for drawing log series   ********************************/ 
    
    onClickLogSeries(){
        this.currState = "logSeries";
        this.drawSeries();
    }
    
    updateSelectedName(name: string) {
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
                name: "agent" + this.logSeries[i].agentid.toString(),
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

    /********************************* functions for statistics information  ************************************/
    
    onClickStatistics() {
        this.currState = "statistics";
        this.drawStatistics();
    }


    updateSelectedBehType(beh: string) {
        this.selectedBehType = beh;
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


    //display Statistics infomation
    drawStatistics() {
        if (this.agentStats !== -1) {
            const methods: string[] = ["max", "min", "count", "average"];
            this.loggerService.getBehavior(this.selectedMASID.toString(), this.agentStats.toString(),
            this.selectedBehType, this.searchStartTime, this.searchEndTime).subscribe( (res: any) => {
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

/********************************* functions for heatmap  ************************************/
    onClickHeatmap() {
        this.currState = "heatmap";
        this.drawHeatmap();
    }

    updatePartition(value: string) {
        if (value === "Yes") {
            this.autoPartition = true;
        } else {
            this.autoPartition = false;
        }
    }

    getPartitionNum(value :string) {
        const num = parseInt(value)
        if (num !== NaN) {
            this.partitionNum = num;
        }
    }

    /************ convert the count of msg communication to color light ***************/
    
    // form hsl(240, 100%, 90%) to hsl(240, 100%, 50%)
    autoConvertColor(values: number[]): string[] {
        const maxVal = Math.max(...values);
        const minVal = Math.min(...values);
        let arrayColor: string[] = [];
        if (maxVal === minVal) {
            arrayColor = Array(values.length).fill("hsl(240, 100%, 70%)")
        } else {
            for (const val of values) {
                let percent: number = 100 -(val - minVal) / (maxVal - minVal) * 40 - 10;
                percent = Math.trunc(percent);
                arrayColor.push("hsl(240, 100%, " + percent.toString() + "%)");
            }
        }
        return arrayColor;   
    }

    

    getColorPartitionEle(): string[] {
        // form hsl(240, 100%, 50%) to hsl(240, 100%, 90%)
        let colors: string[] = [];
        for (let i = 0; i < this.partitionNum; i++) {
            let percent: number = 100 - i / (this.partitionNum - 1) * 40 - 10;
            percent = Math.trunc(percent);
            colors.push("hsl(240, 100%, " + percent.toString() + "%)");
        }
        return colors;
    }

    getColorLegendTexts(values: number[]): string[] {
        const quo: number = Math.floor(values.length / this.partitionNum)
        const remainder: number = values.length % this.partitionNum
        let res: string[] = []
        for (let i = 0; i < remainder; i++) {
            if (values[i * (quo + 1)] === values[(i + 1) * (quo + 1) - 1]) {
                res.push(values[i * (quo + 1)].toString()) 
            } else {
                res.push( values[i * (quo + 1)].toString() + "-" + 
                values[(i + 1) * (quo + 1) - 1].toString())
            }
        }   
        for (let i = remainder; i < this.partitionNum; i++) {
            if (values[i * quo + remainder] === values[(i + 1) * quo + remainder - 1]) {
                res.push(values[i * quo + remainder].toString())
            } else {
                res.push(values[i * quo + remainder].toString() + "-" + 
                values[(i + 1) * quo + remainder - 1].toString())
            }
        }   
        return res; 
    }

    manualConvertColor(idx: number, len: number): number {
        const quo = Math.floor(len / this.partitionNum)
        const remainder = len % this.partitionNum
        if (Math.floor(idx / (quo + 1)) < remainder) {
            return Math.floor(idx / (quo + 1));
        } else {
            return Math.floor((idx - remainder) / quo);
        }
    }


    onEnterGrid(i: number){
        this.popoverFrequency = {};
        this.popoverFrequency.x = this.grids[i].x;
        this.popoverFrequency.y = this.grids[i].y;
        this.popoverFrequency.value = this.grids[i].value;
        this.grids[i].color = "hsl(0, 70.8%, 66.5%)";
    }
    onLeaveGrid(i: number){
        this.grids[i].color = this.gridColors[i];
    }

    drawHeatmap() {
        this.loggerService.getMsgHeatmap(this.selectedMASID.toString(), this.searchStartTime, this.searchEndTime).subscribe( (res: any) => {
            let values: number[] = [];
            for (let i = 0; i < res.length; i++) {
                values.push(parseInt(res[i].split("-")[2]))
            }

            if (this.autoPartition) {
                this.gridColors = this.autoConvertColor(values);
                this.colorPartitionEle = [];
                this.colorLegendTexts = [];
                this.grids = [];
                for (let i = 0; i < res.length; i++) {
                    this.grids.push({
                        x: res[i].split("-")[0],
                        y: res[i].split("-")[1],
                        value: res[i].split("-")[2],
                        color: this.gridColors[i]
                    })
                }
            } else {
                this.grids = [];
                this.gridColors = [];
                const sortedSet: number[] = [...(new Set(values))].sort();
                let mapIdx = new Map();
                sortedSet.forEach((ele, idx) => {
                    mapIdx.set(ele, idx);
                })
                if (sortedSet.length < this.partitionNum) {
                    this.partitionNum = sortedSet.length;
                }
                this.colorPartitionEle = this.getColorPartitionEle();
                this.colorLegendTexts = this.getColorLegendTexts(sortedSet);
                console.log(this.colorLegendTexts)
                console.log(this.colorPartitionEle)
                for (const item of res) {
                    const idx = mapIdx.get(parseInt(item.split("-")[2]));
                    const color: string = this.colorPartitionEle[this.manualConvertColor(idx, sortedSet.length)];
                    this.gridColors.push(color)
                    this.grids.push({
                        x: item.split("-")[0],
                        y: item.split("-")[1],
                        value: item.split("-")[2],
                        color: color
                    })
                }
            }
        })
    }
    




}
