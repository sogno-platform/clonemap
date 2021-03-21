import { Component, OnInit } from '@angular/core';
import { MasService} from 'src/app/services/mas.service'
import { ActivatedRoute, Params} from '@angular/router';

@Component({
  selector: 'app-ams',
  templateUrl: './ams.component.html',
  styleUrls: ['./ams.component.css']
})
export class AMSComponent implements OnInit {

    MASs
    selectedMasId: number = -1;
    selectedMAS

    constructor(
        private masService: MasService,
        private route: ActivatedRoute,
    ) { }

    ngOnInit() {
        // get the information for the sidebar
        this.selectedMasId = -1;
        this.masService.getMAS().subscribe((MASs: any) => {
            if (MASs === null) {
                    this.MASs = [];
            } else {
                this.MASs = MASs;
                }   
            });


        // get the concrete content of the selected MAS
        this.route.params.subscribe(
            (params: Params) => {
                if (params.masId) {
                    this.selectedMasId = params.masId;
                    this.masService.getMASById(params.masId).subscribe((selectedMAS) => {
                        this.selectedMAS = selectedMAS;
                    });
                } else {
                    console.log("No masId");
                }
            });
 
    }

        
    

}
