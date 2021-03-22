import { Component, OnInit } from '@angular/core';
import { DfService} from "src/app/services/df.service";
import { MasService} from "src/app/services/mas.service";
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { ActivatedRoute, Params } from '@angular/router';

@Component({
  selector: 'app-df',
  templateUrl: './df.component.html',
  styleUrls: ['./df.component.css']
})
export class DFComponent implements OnInit {
    selectedMASId:number = -1;
    MASs = null;
    alive: boolean = true;
    fileToUpload: File = null;
    display: string = "";
    filename: string = "Choose a file...";
    searched_results;
    constructor(
        private dfService: DfService,
        private masService: MasService,
        private modalService: NgbModal,
        private route: ActivatedRoute
        ) { }

    ngOnInit() {
        this.dfService.getAlive().subscribe( (res: any) => {
            this.alive = res.df;
        }, 
        error => {
            console.log(error);
        });

        // update the sidebar
        this.masService.getMAS().subscribe((MASs: any) => {
            if (MASs === null) {
                console.log(status);
                this.MASs = [];
            } else {
                this.MASs = MASs
            } 
            }, 
            err => {
                console.log(err)  
            }
        );

        this.route.params.subscribe(
            (params: Params) => {
                if (params.masId) {
                    this.selectedMASId = params.masId;
                    this.dfService.getAllSvcs(this.selectedMASId.toString()).subscribe( res => {
                        this.searched_results = res;     
                    })       
                } else {
                    console.log("No masId");
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

    onCreateSVC() {
        const result = JSON.parse(this.display);
        this.dfService.createSvc(this.selectedMASId.toString(),result).subscribe(
            (res) => {
            console.log("success");
            console.log(res);
            this.modalService.dismissAll("uploaded");
            },
            error => {
                console.log(error);
            }
        );
    }

    onSearchSvcs(desc:string, nodeid:string, dist:string) {
        let masid: string = this.selectedMASId.toString();
    
        if (desc == "" && nodeid == "" && dist == "") {
            this.dfService.getAllSvcs(masid).subscribe( res => {
                this.searched_results = res;
                console.log(res);               
            },
            err => console.log(err)
            )
        }

        else if (desc !== "" && nodeid == "" && dist == "") {
            this.dfService.searchSvc(masid, desc).subscribe( res => {
                this.searched_results = res; 
            },
            err => console.log(err)
            )
        }

        else if (desc != "" && nodeid != "" && dist != "") {
            this.dfService.searchSvcWithinDis(masid, desc, nodeid, dist).subscribe( 
                res => {
                    this.searched_results = res;
                },
                err => {
                    console.log(err);
                });
            }
        else {
            this.searched_results = null;
        }
    }
}
