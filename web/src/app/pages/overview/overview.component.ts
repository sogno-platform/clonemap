import { Component, OnInit } from '@angular/core';
import { MAS } from 'src/app/models/MAS.model'
import { MasService } from 'src/app/services/mas.service';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';

@Component({
  selector: 'app-overview',
  templateUrl: './overview.component.html',
  styleUrls: ['./overview.component.css']
})

export class OverviewComponent implements OnInit {
  
    MASs: MAS[];
    fileToUpload: File = null;
    display: string = "";
    filename: string = "Choose a file...";

    constructor(
        private masService: MasService,
        private modalService: NgbModal
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

    openLg(content) {
        this.modalService.open(content, { size: 'lg', centered: true });
    }

    handleFileInput(files: FileList) {
        if (files.length <= 0) {
            return false;
        }
        this.fileToUpload = files.item(0);
        let fr = new FileReader();
        fr.onload = () => {
            this.display = fr.result.toString();
            this.filename = this.fileToUpload.name;
        }
        fr.readAsText(this.fileToUpload);
    }
    

    onCreateMAS() {
        const result = JSON.parse(this.display);
        this.masService.createMAP(result).subscribe(
            (response) => {
            console.log(response);
            this.modalService.dismissAll("uploaded");
            },
            error => {
                console.log(error);
            }
        );
    }


    onDeleteMAS(id: number) {
        console.log(id);
        this.masService.deleteMASById(id).subscribe(
            (res: any) => {
                console.log(res);
            },
            (err) => {
                console.log(err);
            }
        );
    }

 


}
