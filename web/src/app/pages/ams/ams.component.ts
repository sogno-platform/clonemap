import { Component, OnInit } from '@angular/core';
import { MasService} from 'src/app/services/mas.service'
import { ActivatedRoute, Params} from '@angular/router';
import { MAS } from 'src/app/models/MAS.model';
import { MASItem} from 'src/app/models/MAS-item.model'

@Component({
  selector: 'app-ams',
  templateUrl: './ams.component.html',
  styleUrls: ['./ams.component.css']
})
export class AMSComponent implements OnInit {

    MASs: MAS[];  
    selectedMasId: number = -1;
    selectedMAS: MASItem;

    constructor(
        private masService: MasService,
        private route: ActivatedRoute,
    ) { }

    ngOnInit() {
        // get the information for the sidebar
        this.selectedMasId = -1;
        this.masService.getMAS().subscribe((MASs: MAS[]) => {
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
                    this.masService.getMASById(params.masId).subscribe((selectedMAS: MASItem) => {
                        this.selectedMAS = selectedMAS;
                    });
                } else {
                    console.log("No masId");
                }
            });
 
    }

        
    

}
