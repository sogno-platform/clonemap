import { Component, OnInit } from '@angular/core';
import { MasService} from 'src/app/services/mas.service';
import { ActivatedRoute, Params, Router, Event, NavigationEnd} from '@angular/router';


@Component({
  selector: 'app-sidebar',
  templateUrl: './sidebar.component.html',
  styleUrls: ['./sidebar.component.scss']
})
export class SidebarComponent implements OnInit {
  selectedMASID: number = -1;
  MASID: number[] = []; 
  active:string = "overview";
  constructor(
    private masService: MasService,
    private route: ActivatedRoute,
    private router: Router,
  ) { 
    this.router.events.subscribe((event: Event) => {
      if (event instanceof NavigationEnd) {
          const nav: string = this.router.url.split("/")[1];
          const navbar: string[] = ["overview", "ams", "log", "logseries", "stats", "heatmap", "df"];
          this.selectedMASID = Number(this.router.url.split("/")[2]);
           if (navbar.includes(nav) ) {
              if (nav === "overview") {
                this.active = "ams"
              } else {
                this.active = nav;
              }
          }
      }
      });
  }

  ngOnInit(): void {
    // get the information for the sidebar
    this.masService.getMAS().subscribe((MASs: any) => {
      this.MASID = []
      if (MASs !== null) {
          for (let MAS of MASs) {
              if (MAS.status.code != 5) {
                  this.MASID.push(MAS.id)
              }
          }
      } 
    });
  }
}


