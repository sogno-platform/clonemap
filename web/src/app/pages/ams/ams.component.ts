import { Component, OnInit } from '@angular/core';
import {  DefaultAMSService } from 'src/app/openapi-services/ams';
import { ActivatedRoute, Params} from '@angular/router';
import { HttpClient } from '@angular/common/http';


@Component({
  selector: 'app-ams',
  templateUrl: './ams.component.html',
  styleUrls: ['./ams.component.scss']
})
export class AMSComponent implements OnInit {

    MASID: number[] = [];
    MASConfig = [];
    configColumns: string[] = ['name', 'value'];
    containerInfo = [];
    containerColumns: string[] = ['id','image', 'agencies'];
    agentsInfo = [];
    agentsColumns: string[] = ['id','name', 'type','agency'];    
    selectedMASID: number = -1;
    selectedMAS: any = null;
    q: number = 1;
    data: any;



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

                        console.log(this.containerInfo);
                    
        
                    for (const item of selectedMAS.agents.instances) {
                        this.agentsInfo.push({
                            id: item.id,
                            name: item.spec.name,
                            type: item.spec.type,
                            agency: item.address.agency
                        })
                    }
     

                        
                    });
                } else {
                    console.log("No MASID");
                }
        });
    }


}

