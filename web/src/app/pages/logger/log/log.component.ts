import { Component, OnInit, ViewChild} from '@angular/core';
import { LoggerService} from 'src/app/services/logger.service'
import { LogMessage } from 'src/app/models/log.model';
import { forkJoin, Observable} from 'rxjs';
import { Router, Event, NavigationEnd } from '@angular/router';


@Component({
  selector: 'app-log',
  templateUrl: './log.component.html',
  styleUrls: ['./log.component.scss']
})
export class LogComponent implements OnInit {

    alive: boolean = true;
    selectedMASID: number = -1;

    searchStartTime: string = "20210301000000";
    searchEndTime: string = "20211231000000";


    isTopicSelected: boolean[] = [true, true, true, true, true, true];
    topics: string[] = ["error", "debug", "msg", "status", "app" ];
    width: number = 1500;
    height: number = 2000;
    boxWidth: number = 100;
    boxHeight: number = 50;
    logBoxWidth: number = 60;
    logBoxHeight: number = 25;
    interval: number;
    timeline: any = [];
    agentBox: any = [];
    logBoxes: any = [];
    communications: any = [];
    texts = [];  
    popoverContent: string[] = ["This is the content of the popover"];
    logs: LogMessage[] = [];
 
    constructor(
        private loggerService: LoggerService,
        private router: Router
    ) {
        this.router.events.subscribe((event: Event) => {
            if (event instanceof NavigationEnd) {
                this.selectedMASID = Number(this.router.url.split("/")[2]);
            }
            });
        }

    ngOnInit(): void {} 
    


    /********************************** common functions ************************************/
    convertDate(date: Date): string {
        let res: string = date.getFullYear().toString();
        res += (date.getMonth() + 1) < 10 ? "0" + (date.getMonth() + 1).toString() : (date.getMonth() + 1).toString()
        res += date.getDate()  < 10 ? "0" + date.getDate().toString() : date.getDate().toString()
        return res;
    }

    generateScaledDates(dates: Date[]) :number[]{       
        // find the date differences
        let datesInterval: number[] = [0];
        for (let i = 1; i < dates.length; i++) {
            datesInterval.push(Math.floor(dates[i].getTime()/1000) - Math.floor(dates[i-1].getTime()/1000));
        }

        // find the maximum and minimum interva
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

    onClickSearchButton(
        selectedAgent,
        selectedStartDate,
        selectedEndDate,
        selectedStartTime,
        selectedEndTime
        ) {
        const startDate: string = this.convertDate(selectedStartDate);
        const endDate:string =  this.convertDate(selectedEndDate);
        const searchStartTime = startDate + selectedStartTime.replace(":", "") + "00";
        const searchEndTime = endDate + selectedEndTime.replace(":", "") + "59";
        this.drawLogs(selectedAgent, searchStartTime, searchEndTime);
  
    }

    downloadData(blobConfig: Blob, filename: string) {

        // Convert Blob to URL
        const blobUrl = URL.createObjectURL(blobConfig);

        // Create an a element with blobl URL
        const anchor = document.createElement('a');
        anchor.href = blobUrl;
        anchor.target = "_blank";
        anchor.download = filename;

        // Auto click on a element, trigger the file download
        anchor.click();

        URL.revokeObjectURL(blobUrl);

    }

    onDownloadLogs(selectedAgent,
        selectedStartDate,
        selectedEndDate,
        selectedStartTime,
        selectedEndTime
        ) {
        const startDate: string = this.convertDate(selectedStartDate);
        const endDate:string =  this.convertDate(selectedEndDate);
        const searchStartTime = startDate + selectedStartTime.replace(":", "") + "00";
        const searchEndTime = endDate + selectedEndTime.replace(":", "") + "59";
        // configs object
        this.multiLogs(selectedAgent, searchStartTime, searchEndTime).subscribe( logss => {
            let config: LogMessage[] = [];
            for (let logs of logss) {
                if (logs !== null) {
                    for (let log of logs) {
                        config.push(log);
                    }
                }
            }
            // Convert object to Blob
            const blobConfig = new Blob(
                [ JSON.stringify(config) ], 
                { type: 'text/json;charset=utf-8' }
            )
            this.downloadData(blobConfig, "log.json")
        })
    
    }

    /********************************* functions for drawing logs  ************************************/

    onToggleTopic(i: number) {
        this.isTopicSelected[i] = ! this.isTopicSelected[i];
        this.updateScaledDates();
    }

    drawLogs(selectedAgent,searchStartTime, searchEndTime
        ) {
        this.logs = [];     
        this.multiLogs(selectedAgent, searchStartTime, searchEndTime).subscribe( logss => {
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
                return date1.getTime() - date2.getTime();
            });
            this.drawAllElements(selectedAgent);
         })
    }

    multiLogs(selectedAgent, searchStartTime, searchEndTime): Observable<any[]> {
        let res = [];
        for (let id of selectedAgent) {
            for (let topic of this.topics) {
                res.push(this.loggerService.getLogsInRange(this.selectedMASID.toString(),
                id.toString(), topic, searchStartTime, searchEndTime));
            }
        }
        return forkJoin(res);
    } 

    drawAgentBox(selectedAgent: number[]) {
        this.agentBox = [];
        this.texts = [];
        this.interval = 1 / (1 + selectedAgent.length) * this.width;

        for (let i=0; i < selectedAgent.length; i++) {
            const X: number = (i+1) * this.interval;
            // plot the agent box
            this.agentBox.push({x: X - this.boxWidth / 2, y: 200 - this.boxHeight});
            this.texts.push({
                x: X - this.boxWidth * 5 / 12, 
                y: 200 - this.boxHeight / 3,
                textID: selectedAgent[i],
            })
        }
    }

    drawTimeline(selectedAgent) {
        this.timeline = []
        for (let i = 0; i < selectedAgent.length; i++) {
            let X = (i+1) * this.interval;
            this.timeline.push({x1: X, y1:200, x2: X, y2: this.height })  
        }
    }

    drawScaledDates(scaledDates: number[], selectedAgent: number[]) {
        this.logBoxes = [];
        this.communications = [];
        let agentLogs = []
        for (let i = 0; i < selectedAgent.length; i++) {
            agentLogs[i] = [];
        }
        for (let i = 0; i < scaledDates.length; i++) {
            let currMsg = this.logs[i];
            let agentIndex = selectedAgent.indexOf(currMsg.agentid);
            let idx = selectedAgent.indexOf(currMsg.agentid) + 1;
            let r_x = 0;
            let r_y = 0;
            if (currMsg.topic === "msg" && (currMsg.msg === "ACL receive" || currMsg.msg === "MQTT receive")) {
                r_x = 7;
                r_y = 7;
            }
            this.logBoxes.push({
                x: this.interval *idx - this.logBoxWidth / 2, 
                y: 400 + this.logBoxHeight * scaledDates[i] * 1.1,
                width: this.logBoxWidth,
                topic: currMsg.topic,
                rx: r_x,
                ry: r_y,
                timestamp: currMsg.timestamp,
                msg: currMsg.msg,
                data: currMsg.data,
                hidden: !this.isTopicSelected[this.topics.indexOf(currMsg.topic)],
                
            });
            agentLogs[agentIndex].push(i);
            if (currMsg.topic === "msg" && currMsg.msg ==="ACL send"){
                const data = this.logs[i].data.split(";");
                const sender = Number(data[0].split(" ")[1]);
                const senderIdx = selectedAgent.indexOf(sender) + 1;
                const receiver = Number(data[1].split(" ")[1]);
                const receiverIdx = selectedAgent.indexOf(receiver) + 1;
                const direction = (senderIdx < receiverIdx) ? 1 : -1;
                if (selectedAgent.includes(receiver) && selectedAgent.includes(sender)) {
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
        // display logs for the same agent at the same time in parallel
        for (let i = 0; i < agentLogs.length; i++) {
            let parallelLogs = 1;
            for (let j = 0; j < agentLogs[i].length; j++) {
                if (j > 0) {
                    if (scaledDates[agentLogs[i][j]] === scaledDates[agentLogs[i][j-1]]) {
                        parallelLogs++
                    } else {
                        parallelLogs = 1;
                    }
                    if (parallelLogs > 1) {
                        for (let k = 0; k < parallelLogs; k++) {
                            let width = Math.round(this.logBoxWidth / parallelLogs);
                            let xPos = this.interval *(i+1) - this.logBoxWidth / 2 + width * k;
                            this.logBoxes[agentLogs[i][j-parallelLogs+1+k]].x = xPos;
                            this.logBoxes[agentLogs[i][j-parallelLogs+1+k]].width = width;
                        }
                    }
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
                
    drawAllElements(selectedAgent: number[]) {
        this.drawAgentBox(selectedAgent);
        let dates: Date[] = []
        for (let i = 0; i < this.logs.length; i++) {
            let date = new Date(this.logs[i].timestamp)
            dates.push(date)
        }
        
        let scaledDates: number[] = this.generateScaledDates(dates);
        this.height = 800 + this.logBoxHeight * scaledDates[scaledDates.length-1];
        this.drawScaledDates(scaledDates, selectedAgent);
        this.drawTimeline(selectedAgent);
    }

    onChangePopoverContent(i) {
        if ("data" in this.logs[i]) {
            if (this.logs[i].msg === "ACL send" || this.logs[i].msg === "ACL receive") {
                this.popoverContent = this.logs[i].data.split(";");
                this.popoverContent[2] = this.popoverContent[2].split(".")[0]; 
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


}
