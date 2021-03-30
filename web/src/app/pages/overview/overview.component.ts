import { Component, OnInit } from '@angular/core';
import { MasService } from 'src/app/services/mas.service';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { Router } from '@angular/router'

@Component({
  selector: 'app-overview',
  templateUrl: './overview.component.html',
  styleUrls: ['./overview.component.css']
})

export class OverviewComponent implements OnInit {
  
    MASs = null;
    MASsDisplay = null;
    fileToUpload: File = null;
    display: string = "";
    filename: string = "Choose a file...";
    status: string ="Connecting......";
    constructor(
        private masService: MasService,
        private modalService: NgbModal,
        private router: Router

    ) {} 
    
    ngOnInit() {
    
        this.masService.getMAS().subscribe((MASs: any) => {
            if (MASs === null) {
                this.status = "Currently no agencies, upload one......";
                this.MASs = [];
                this.MASsDisplay = [];
                console.log(this.MASs);
                
            } else {
                this.MASs = MASs;
                this.MASsDisplay = MASs;
            }
            },
            err => {
                this.status = "The CloneMAP platform is not connected"
                console.log(err)  
            }
        );


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

    onUpdateContent(content:string) {
        this.display=content;
    }
    

    onCreateMAS() {
        const result = JSON.parse(this.display);
        this.masService.createMAS(result).subscribe(
            (response) => {
            console.log(response);
            this.modalService.dismissAll("uploaded");
            },
            error => {
                console.log(error);
            }
        );
    }

    onDeleteMAS(id: string) {
        console.log(id);
        this.masService.deleteMASById(id).subscribe(
            (res: any) => {
                console.log(res);
                this.router.navigate['/overview']
            },
            (err) => {
                console.log(err);
            }
        );
    }

    onSearchMAS(id: string) {
        console.log(id);
        if (id == "") {
            this.MASsDisplay = this.MASs;
            return;
        }
        this.masService.getMASById(id).subscribe(
            (res: any) => {
                console.log(res);
                this.MASsDisplay = [res];
                this.router.navigate['/overview']
            },
            (err) => {
                //this.MASs = []
                console.log(err);
                this.MASsDisplay=[]
            }
        );
    }



}
