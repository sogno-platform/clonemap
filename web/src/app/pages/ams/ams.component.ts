import { Component, OnInit } from '@angular/core';
import { MasService} from 'src/app/services/mas.service'
import { ActivatedRoute, Params} from '@angular/router';

@Component({
  selector: 'app-ams',
  templateUrl: './ams.component.html',
  styleUrls: ['./ams.component.css']
})
export class AMSComponent implements OnInit {

    MASID: number[] = [];
    selectedMASID: number = -1;
    selectedMAS: any = null;

    constructor(
        private masService: MasService,
        private route: ActivatedRoute,
    ) { }

    ngOnInit() {

        // get the information for the sidebar
        this.masService.getMAS().subscribe((MASs: any) => {
            if (MASs !== null) {
                this.MASID = MASs.map( MAS => MAS.id);
                console.log(this.MASID);
            }   
        });

        // get the concrete content of the selected MAS
        this.route.params.subscribe(
            (params: Params) => {
                if (params.masid) {
                    this.selectedMASID = params.masid;
                    this.masService.getMASById(params.masid).subscribe((selectedMAS: any) => {
                        this.selectedMAS = selectedMAS;
                    });
                } else {
                    console.log("No MASID");
                }
        });
    }

}
