import { Component, OnInit, Renderer2, Inject } from '@angular/core';
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
    selectedMASID:number = -1;
    MASs = null;
    alive: boolean = true;
    fileToUpload: File = null;
    display: string = "";
    filename: string = "Choose a file...";
    collapsed: boolean[] = [];
    searched_results;
    constructor(
        private dfService: DfService,
        private masService: MasService,
        private modalService: NgbModal,
        private route: ActivatedRoute,
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
            this.MASs = MASs;
            }, 
            err => {
                console.log(err)  
            }
        );

        this.route.params.subscribe(
            (params: Params) => {
                if (params.masid) {
                    this.selectedMASID = params.masid;
                    this.dfService.getAllSvcs(this.selectedMASID.toString()).subscribe( (res:any) => {
                        this.searched_results = res;   
                        for (let i = 0; i < res.length; i++) {
                            this.collapsed.push(false);
                        }  
                    })       
                } else {
                    console.log("No masid");
                }
                
            });
    }

    // functions for update the services
    openLg(content) {
        this.modalService.open(content, { size: 'lg', centered: true });
    }

    onUpdateContent(content:string) {
        this.display=content;
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

    onToggleCollapsed(i: number) {
        this.collapsed[i] = !this.collapsed[i];
    }



    onCreateSVC() {
        const result = JSON.parse(this.display);
        this.dfService.createSvc(this.selectedMASID.toString(),result).subscribe(
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
        let masid: string = this.selectedMASID.toString();
    
        if (desc === "" && nodeid === "" && dist === "") {
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

        else if (desc !== "" && nodeid !== "" && dist !== "") {
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

        this.collapsed = [];
        for (let i = 0; i < this.searched_results.length; i++) {
            this.collapsed.push(true);
        }
    }
}
