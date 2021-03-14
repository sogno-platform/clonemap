import { Component, OnInit } from '@angular/core';
import { MAS } from 'src/app/models/MAS.model'
import { MasService } from 'src/app/services/mas.service';

@Component({
  selector: 'app-overview',
  templateUrl: './overview.component.html',
  styleUrls: ['./overview.component.css']
})

export class OverviewComponent implements OnInit {
  
    MASs: MAS[];

    constructor(
        private masService: MasService
    ) {} 
    
    ngOnInit() {
        this.masService.getMAS().subscribe((MASs: MAS[]) => {
            if (MASs === null) {
                this.MASs = [];
            } else {
                this.MASs = MASs;
                }   
        });
    }

}
