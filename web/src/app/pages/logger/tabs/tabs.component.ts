import { Component, OnInit } from '@angular/core';
import { Router , Event, NavigationEnd} from '@angular/router';

@Component({
  selector: 'app-tabs',
  templateUrl: './tabs.component.html',
  styleUrls: ['./tabs.component.scss']
})
export class TabsComponent implements OnInit {

	currState: string = "log";
	masid: string = '1';

	constructor(private router: Router) { 
		this.router.events.subscribe((event: Event) => {
		if (event instanceof NavigationEnd) {
			const tab: string = this.router.url.split("/")[1];
			this.masid = this.router.url.split("/")[2];
			const tabs: string[] = ["log", "logseries", "stats", "heatmap"];
			if (tabs.includes(tab) ) {
				this.currState = tab;
			}
		}
		});
	}
	
	ngOnInit(): void {
	}

	onClickLog() {
		this.router.navigate(['/log/' + this.masid]);
		this.currState = "log";
	}

	onClickLogSeries() {
		this.router.navigate(['/logseries/' + this.masid]);
		this.currState = "logSeries";
	}

	onClickStats() {
		this.router.navigate(['/stats/' + this.masid]);
		this.currState = "stats";
	}

	onClickHeatmap() {
		this.router.navigate(['/heatmap/' + this.masid]);
		this.currState = "heatmap";
	}
	

}
