import { Component, OnInit, ViewChild, AfterViewInit } from '@angular/core';
import {  DefaultAMSService } from 'src/app/openapi-services/ams';
import { ActivatedRoute, Params} from '@angular/router';
import { MatPaginator } from '@angular/material/paginator';
import { HttpClient } from '@angular/common/http';
import { MatSort, SortDirection} from '@angular/material/sort';
import { merge, Observable, of as observableOf} from 'rxjs';
import { catchError, map, startWith, switchMap } from 'rxjs/operators';

@Component({
  selector: 'app-ams',
  templateUrl: './ams.component.html',
  styleUrls: ['./ams.component.scss']
})
export class AMSComponent implements OnInit, AfterViewInit {

    MASID: number[] = [];
    MASConfig = [];
    configColumns: string[] = ['name', 'value'];
    containerInfo = [];
    containerColumns: string[] = ['id','image', 'agencies'];
    selectedMASID: number = -1;
    selectedMAS: any = null;
    q: number = 1;
    data: any;

    @ViewChild(MatPaginator) paginator: MatPaginator;
    @ViewChild(MatSort) sort: MatSort;

    resultsLength = 0;
    isLoadingResults = true;
    isRateLimitReached = false;    

    constructor(
        private amsService: DefaultAMSService,
        private route: ActivatedRoute,
        private http: HttpClient
    ) { }


    ngOnInit() {

        // get the concrete content of the selected MAS
        this.route.params.subscribe(
            (params: Params) => {
                if (params.masid) {
                    this.selectedMASID = params.masid;
                    this.amsService.getOneMAS(params.masid).subscribe((selectedMAS: any) => {
                        this.selectedMAS = selectedMAS;
                        
                        // create config table
                        this.MASConfig.push({
                            name: "ID",
                            value: this.selectedMAS.id
                        })
                        this.MASConfig.push({
                            name: "Name",
                            value: this.selectedMAS.config.name
                        })
                        this.MASConfig.push({
                            name: "Agents per agency",
                            value: this.selectedMAS.config.agentsperagency
                        })
                        this.MASConfig.push({
                            name: "DF",
                            value: this.selectedMAS.config.df.active
                        })
                        this.MASConfig.push({
                            name: "Logging",
                            value: this.selectedMAS.config.logger.active
                        })
                        this.MASConfig.push({
                            name: "MQTT",
                            value: this.selectedMAS.config.mqtt.active
                        })

                    // create container table 
                        for (const item of selectedMAS.imagegroups.instances) {
                            this.containerInfo.push({
                                id: item.id,
                                image: item.config.image,
                                agencies: item.agencies.counter

                            })
                        }
                        
                    });
                } else {
                    console.log("No MASID");
                }
        });
    }

    ngAfterViewInit() {
  

        // If the user changes the sort order, reset back to the first page.
        this.sort.sortChange.subscribe(() => this.paginator.pageIndex = 0);
    
        merge(this.sort.sortChange, this.paginator.page)
          .pipe(
            startWith({}),
            switchMap(() => {
              this.isLoadingResults = true;
              return this.amsService.getOneMAS(this.selectedMASID)
                .pipe(catchError(() => observableOf(null)));
            }),
            map(data => {
              // Flip flag to show that loading has finished.
              this.isLoadingResults = false;
              this.isRateLimitReached = data === null;
    
              if (data === null) {
                return [];
              }
    
              // Only refresh the result length if there is new data. In case of rate
              // limit errors, we do not want to reset the paginator to zero, as that
              // would prevent users from re-triggering requests.
              this.resultsLength = data.total_count;
              return data.items;
              console.log(data);
            })
          ).subscribe(data => this.data = data);
    }

}

