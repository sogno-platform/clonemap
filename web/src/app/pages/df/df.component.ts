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
    selectedMASId:number = -1;
    MASs = null;
    alive: boolean = true;
    fileToUpload: File = null;
    display: string = "";
    filename: string = "Choose a file...";
    searched_results;
    curr_state = "list";
    dashed_connection = [];
    services = [];
    comms = [[1, 2], [2, 1], [1, 3], [5, 4], [7, 3], [8,1], [4, 6], [3, 6], [1, 7], [1, 3]]; // fake communications
    margin: number = 50;
    interactions = [];
    rects = [];
    texts = [];
    gap_svc: number = 100;
    gap_comm: number = 50;
    constructor(
        private dfService: DfService,
        private masService: MasService,
        private modalService: NgbModal,
        private route: ActivatedRoute,
        private renderer2: Renderer2
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
                if (params.masId) {
                    this.selectedMASId = params.masId;
                    this.dfService.getAllSvcs(this.selectedMASId.toString()).subscribe( res => {
                        this.searched_results = res;     
                    })       
                } else {
                    console.log("No masId");
                }
                
            });

        for (let i = 1; i <= 8; i++) {
            let x = i * this.gap_svc;
            this.dashed_connection.push({x1: x, y1: 150, x2:x, y2:200});
            this.services.push({x1:x, y1: 200, x2:x, y2: 900,});
            this.rects.push({x: 60+(i-1)*100, y: 70});
            this.texts.push({x: 90+(i-1)*100, y: 110})
        }
        // create the interaction pairs
        for (let i = 0; i < this.comms.length; i++) {
            const top = 250;
            this.interactions.push({x1:this.comms[i][0] * this.gap_svc, y1: top + i * this.gap_comm, 
                x2:this.comms[i][1] * this.gap_svc, y2: top + i * this.gap_comm})
        }
        
    }

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
    }

    onClickList() {
        this.curr_state = "list";
        console.log(this.curr_state);
    }

    onClickUML() {
        this.curr_state = "UML";
        console.log(this.curr_state);
    }
}
