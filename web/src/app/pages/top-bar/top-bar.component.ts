import { Component, OnInit } from '@angular/core';
import { Router, Event, NavigationEnd } from '@angular/router';


@Component({
  selector: 'app-top-bar',
  templateUrl: './top-bar.component.html',
  styleUrls: ['./top-bar.component.css']
})
export class TopBarComponent implements OnInit {
    active = 'overview';

    constructor(private router: Router) { 
        this.router.events.subscribe((event: Event) => {
        if (event instanceof NavigationEnd) {
            const nav: string = this.router.url.split("/")[1];
            if (nav in ["overview", "ams", "logger", "df"]) {
                this.active = nav;
            }
        }
        });
    }

    ngOnInit(){
    }

}
