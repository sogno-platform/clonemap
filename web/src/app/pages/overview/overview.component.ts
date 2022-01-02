import { Component, OnInit } from '@angular/core';
import { NgbModal, NgbModalConfig } from '@ng-bootstrap/ng-bootstrap';
import { Router, ActivatedRouteSnapshot } from '@angular/router'
import { HttpResponse } from '@angular/common/http';
import { DefaultAMSService } from 'src/app/openapi-services/ams';

@Component({
  selector: 'app-overview',
  templateUrl: './overview.component.html',
  styleUrls: ['./overview.component.scss']
})

export class OverviewComponent implements OnInit {
  
    MASID: number[] = [];
    MASsDisplay: any = [];
    fileToUpload: File = null;
    display: string = "";
    filename: string = "Choose a file...";
    status: string ="Connecting......";
    constructor(
        config: NgbModalConfig,
        private modalService: NgbModal,
        private router: Router,
        private amsService: DefaultAMSService,
    ) {
        config.backdrop = "static";
    } 

    ngOnInit() {
        this.updateMAS();
        this.amsService.getAllMASs().subscribe(res => {
            console.log(res);
        })
    }


    /********************   create new MAS *************************/

    openLg(content) {
        this.modalService.open(content, {centered: true, size:"lg" });
    }

    closeModal() {
        this.modalService.dismissAll();
        this.updateMAS();
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
        this.amsService.createNewMAS(result).subscribe(
            (response) => {
            this.modalService.dismissAll();
            this.updateMAS();
            },
            error => {
                this.modalService.dismissAll();
                this.updateMAS();
                console.log(error);
            }
        );
    }

    updateMAS() {
        this.amsService.getAllMASs().subscribe((MASs: any) => {
            this.MASsDisplay = [];
            if (MASs !== null) {
                for (let MAS of MASs) {
                    // if the MAS is not deleted
                    if (MAS.status.code != 5) {
                        this.MASsDisplay.push(MAS)
                    }
                }
                this.MASID = this.MASsDisplay.map(MAS => MAS.id);
                if (this.MASsDisplay.length === 0) {
                    this.status = "Currently no MASs, create one......";
                } else {
                    this.status = "Connected successfully!"
                }
            } else {
                this.status = "Currently no MASs, create one......";
            }
        }, err => {
                this.status = "The CloneMAP platform is not connected";
            });
    }

    /********************   delete the MAS *************************/
    onDeleteMAS(id: string, deleting, deleted) {
        this.modalService.open(deleting, { size: 'sm', centered: true });
        this.amsService.deleteOneMAS(parseInt(id)).subscribe(
            (res: HttpResponse<any>) => {
                this.modalService.dismissAll();
                this.modalService.open(deleted, { size: 'sm', centered: true });
            },
            (err) => {
                console.log(err)
            }
        );
    }

    onOpenMAS(i: number) {
        this.router.navigate(['/ams', i]);
    }

}
