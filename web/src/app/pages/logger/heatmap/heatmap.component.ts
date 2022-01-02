import { Component, OnInit } from '@angular/core';
import { DefaultLoggerService } from 'src/app/openapi-services/logger';
import { Router, Event, NavigationEnd } from '@angular/router';


@Component({
  selector: 'app-heatmap',
  templateUrl: './heatmap.component.html',
  styleUrls: ['./heatmap.component.scss']
})
export class HeatmapComponent implements OnInit {
    selectedMASID: number = -1;

    gridColors: string[] = [];
    grids: any[] = [];
    gridWidth: number = 4;
    popoverFrequency: any = {};
    autoPartition: boolean = false;
    partitionNum: number = 5;
    colorPartitionEle: string[] = [];
    colorLegendTexts: string[] = [];
    legendWidth = 30;

        
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

    ngOnInit(): void {
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
        this.drawHeatmap(searchStartTime, searchEndTime);
        
    }

/************ convert the count of msg communication to color light ***************/

    // automatically generate heatmap from hsl(240, 100%, 90%) to hsl(240, 100%, 50%)
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

    // generate heatmap with self defined number of partitions
    getColorPartitionEle(): string[] {
        // from hsl(240, 100%, 50%) to hsl(240, 100%, 90%)
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

    drawHeatmap(searchStartTime: string, searchEndTime: string) {
    this.loggerService.getMsgHeatmap(this.selectedMASID, searchStartTime, searchEndTime).subscribe( (res: any) => {
        let values: number[] = [];
        let maxID: number = 0;
        for (let i = 0; i < res.length; i++) {
            values.push(parseInt(res[i].split("-")[2]))
        }

        if (this.autoPartition) {
            this.gridColors = this.autoConvertColor(values);
            this.colorPartitionEle = [];
            this.colorLegendTexts = [];
            this.grids = [];
            for (let i = 0; i < res.length; i++) {
                let x = res[i].split("-")[0]
                let y = res[i].split("-")[1]
                this.grids.push({
                    x: x,
                    y: y,
                    value: res[i].split("-")[2],
                    color: this.gridColors[i]
                })
                if (x > maxID) {
                    maxID = x
                }
                if (y > maxID) {
                    maxID = y
                }
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
            for (const item of res) {
                const idx = mapIdx.get(parseInt(item.split("-")[2]));
                const color: string = this.colorPartitionEle[this.manualConvertColor(idx, sortedSet.length)];
                this.gridColors.push(color)
                let x = item.split("-")[0]
                let y = item.split("-")[1]
                this.grids.push({
                    x: x,
                    y: y,
                    value: item.split("-")[2],
                    color: color
                })
                if (x > maxID) {
                    maxID = x
                }
                if (y > maxID) {
                    maxID = y
                }
            }
        }
        this.gridWidth = 3000 / (maxID+1)
    })
    }

    convertDate(date: Date): string {
        let res: string = date.getFullYear().toString();
        res += (date.getMonth() + 1) < 10 ? "0" + (date.getMonth() + 1).toString() : (date.getMonth() + 1).toString()
        res += date.getDate()  < 10 ? "0" + date.getDate().toString() : date.getDate().toString()
        return res;
    }

}
