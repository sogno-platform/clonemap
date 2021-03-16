import { Component, OnInit } from '@angular/core';
import { DfService} from "src/app/services/df.service"

@Component({
  selector: 'app-df',
  templateUrl: './df.component.html',
  styleUrls: ['./df.component.css']
})
export class DFComponent implements OnInit {

    status: string = "I am not alive";

    constructor(private dfService: DfService) { }

    ngOnInit() {
        this.dfService.getAlive().subscribe( res => {
            this.status =  res.toString();
        }, error => {
            console.log(error);
        });
        
        this.dfService.getAllSvcs(0).subscribe( res => {
            console.log(res);
        }, error => {
            console.log(error);
        })


    }





}
