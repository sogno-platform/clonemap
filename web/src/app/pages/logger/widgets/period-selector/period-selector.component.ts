import { Component, OnInit } from '@angular/core';

@Component({
  selector: 'app-period-selector',
  templateUrl: './period-selector.component.html',
  styleUrls: ['./period-selector.component.scss']
})
export class PeriodSelectorComponent implements OnInit {

  constructor() { }

  ngOnInit(): void {
  }

  selectedStartDate: Date = new Date();
  selectedEndDate: Date = new Date();
  selectedStartTime: string = "00:00";
  selectedEndTime: string = "23:59";
}
