import { Component, OnInit } from '@angular/core';
import { DefaultLoggerService } from 'src/app/openapi-services/logger';
import { Router, Event, NavigationEnd } from '@angular/router';
import { LogSeries, pointSeries } from 'src/app/models/log.model';
import { forkJoin, Observable} from 'rxjs';

@Component({
  selector: 'app-logseries',
  templateUrl: './logseries.component.html',
  styleUrls: ['./logseries.component.scss']
})
export class LogseriesComponent implements OnInit {

    selectedMASID: number = -1;
    gridWidth: number;
    
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

    constructor(
        private loggerService: DefaultLoggerService,
        private router: Router
        ) {
            this.router.events.subscribe((event: Event) => {
                if (event instanceof NavigationEnd) {
                    this.selectedMASID = Number(this.router.url.split("/")[2]);
                }
            });
    }

    ngOnInit(): void {}

    convertDate(date: Date): string {
        let res: string = date.getFullYear().toString();
        res += (date.getMonth() + 1) < 10 ? "0" + (date.getMonth() + 1).toString() : (date.getMonth() + 1).toString()
        res += date.getDate()  < 10 ? "0" + date.getDate().toString() : date.getDate().toString()
        return res;
    }

    downloadData(blobConfig: Blob, filename: string) {
        // Convert Blob to URL
        const blobUrl = URL.createObjectURL(blobConfig);

        // Create an a element with blobl URL
        const anchor = document.createElement('a');
        anchor.href = blobUrl;
        anchor.target = "_blank";
        anchor.download = filename;

        // Auto click on an element, trigger the file download
        anchor.click();

        URL.revokeObjectURL(blobUrl);
    }

    onDownloadLogs(
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
        // configs object
        let config: LogSeries[] = [];
        this.multiSeries(selectedAgent, searchStartTime, searchEndTime).subscribe( logss => {
            for (let logs of logss) {
                if (logs !== null) {
                    for (let log of logs) {
                        config.push(log);
                    }
                }
            }
            let configStr: string = config.map(log => {
                return [log.masid, log.agentid, log.name, log.timestamp, log.value].join(",")
            }).join("\n");
            // Convert object to Blob
            const blobConfig = new Blob(
                [configStr], 
                { type: 'text/csv;charset=utf-8' }
            )
            this.downloadData(blobConfig, "log_series.csv")
        })
    }


    generateScaledDates(dates: Date[]) :number[]{       
        // find the date differences
        let datesInterval: number[] = [0];
        for (let i = 1; i < dates.length; i++) {
            datesInterval.push(Math.round(dates[i-1].getTime()/1000) - Math.round(dates[i].getTime()/1000));
        }

        // find the maximum and minimum interval
        let minDiff: number = Number.MAX_SAFE_INTEGER;
        let maxDiff: number = 0

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
        let curr: number = 0;
        if (maxDiff !== minDiff) {
            for (let i = 0; i < datesInterval.length; i++) {
                if (datesInterval[i] !== 0) {
                    curr = curr +  Math.round(100 * ((5 - 1)  * (datesInterval[i] - minDiff)/(maxDiff - minDiff) + 1)) / 100;
                }
                scaledDates.push(curr); 
            }
        // if maximum difference is the same as the minimum difference, convert every interval to 1 
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
        this.drawSeries(selectedAgent, searchStartTime, searchEndTime );
        
    }

    updateSelectedName(name: string) {
        this.selectedName = name;
    }

    updateNames(selectedAgent: number[]) {
        this.names = [];
        this.multiNames(selectedAgent).subscribe( namess => {
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

    // return the observable of log series names of the selected agent 
    multiNames(selectedAgent: number[]): Observable<any> {
        let res = [];
        for (let id of selectedAgent) {
            res.push(this.loggerService.getLogSeriesNames(this.selectedMASID,
            id));
        }
        return forkJoin(res);      
    }

    // return the observable of log series of the selected names
    multiSeries(selectedAgent: number[], searchStartTime, searchEndTime): Observable<any> {
        let res = [];
        for (let id of selectedAgent) {
            res.push(this.loggerService.getLogSeriesInRange(this.selectedMASID,
            id, this.selectedName, searchStartTime, searchEndTime));
        }
        return forkJoin(res);
    }

    drawSeries(selectedAgent: number[], searchStartTime, searchEndTime) {   
        this.logSeries = [];
        this.multiSeries(selectedAgent, searchStartTime, searchEndTime).subscribe( logss => {
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

    // generate the data for the scatter plot of log series
    drawSeriesHelper() {
    this.datesSeries = [];
    this.bubbleData = [];
    for (let i = 0; i < this.logSeries.length; i++) {
        let date = new Date(this.logSeries[i].timestamp)
        this.datesSeries.push(date)
    }
    this.scaledDatesSeries = this.generateScaledDates(this.datesSeries); 
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
        this.mapAxisDate.set(x, new Date(this.logSeries[i].timestamp).toLocaleString('de-DE',{ hour12: false }) )
        this.bubbleData.push({
            name: "agent" + this.logSeries[i].agentid.toString(),
            series: [point]
        })
    }
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


}
