import { Component, OnInit, Renderer2, Inject } from '@angular/core';
import { DfService} from "src/app/services/df.service";
import { MasService} from "src/app/services/mas.service";
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { ActivatedRoute, Params } from '@angular/router';
import * as cytoscape from 'cytoscape';
import euler from 'cytoscape-euler';
import spread from 'cytoscape-spread';

cytoscape.use(euler);

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
    curr_state: string = "graph";
    disc;
    graph;
    nodes;
    edges;
    agents;
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

    ngAfterViewInit() {
            this.graph = cytoscape({
            container: document.getElementById('graph'),
            elements: [
                // nodes
                { data: { id: 'a' } },
                { data: { id: 'b' } },
                // edges
                {
                  data: {
                    id: 'ab',
                    source: 'a',
                    target: 'b'
                  }
                }
            ],
             
/*             layout: {
                name: 'circle'
            }, */
            style: [
                {
                    selector: 'node',
                    style: {
                        'background-color': "#9696f3",
                        label: 'data(id)'
                    }
                }
            ]
        });

        for (var i = 0; i < 10; i++) {
            this.graph.add({
                data: { id: 'node' + i}
            });
            var source = 'node' + i;
            this.graph.add({
                data: {
                    id: 'edge' + i,
                    source: source,
                    target:(i % 2 == 0 ? 'a' : 'b')
                }
            })
        }

        let defaults = {
            name: 'euler',
          
            // The ideal length of a spring
            // - This acts as a hint for the edge length
            // - The edge length can be longer or shorter if the forces are set to extreme values
            springLength: edge => 80,
          
            // Hooke's law coefficient
            // - The value ranges on [0, 1]
            // - Lower values give looser springs
            // - Higher values give tighter springs
            springCoeff: edge => 0.0008,
          
            // The mass of the node in the physics simulation
            // - The mass affects the gravity node repulsion/attraction
            mass: node => 4,
          
            // Coulomb's law coefficient
            // - Makes the nodes repel each other for negative values
            // - Makes the nodes attract each other for positive values
            gravity: -1.2,
          
            // A force that pulls nodes towards the origin (0, 0)
            // Higher values keep the components less spread out
            pull: 0.001,
          
            // Theta coefficient from Barnes-Hut simulation
            // - Value ranges on [0, 1]
            // - Performance is better with smaller values
            // - Very small values may not create enough force to give a good result
            theta: 0.666,
          
            // Friction / drag coefficient to make the system stabilise over time
            dragCoeff: 0.02,
          
            // When the total of the squared position deltas is less than this value, the simulation ends
            movementThreshold: 1,
          
            // The amount of time passed per tick
            // - Larger values result in faster runtimes but might spread things out too far
            // - Smaller values produce more accurate results
            timeStep: 20,
          
            // The number of ticks per frame for animate:true
            // - A larger value reduces rendering cost but can be jerky
            // - A smaller value increases rendering cost but is smoother
            refresh: 10,
          
            // Whether to animate the layout
            // - true : Animate while the layout is running
            // - false : Just show the end result
            // - 'end' : Animate directly to the end result
            animate: true,
          
            // Animation duration used for animate:'end'
            animationDuration: undefined,
          
            // Easing for animate:'end'
            animationEasing: undefined,
          
            // Maximum iterations and time (in ms) before the layout will bail out
            // - A large value may allow for a better result
            // - A small value may make the layout end prematurely
            // - The layout may stop before this if it has settled
            maxIterations: 1000,
            maxSimulationTime: 4000,
          
            // Prevent the user grabbing nodes during the layout (usually with animate:true)
            ungrabifyWhileSimulating: false,
          
            // Whether to fit the viewport to the repositioned graph
            // true : Fits at end of layout for animate:false or animate:'end'; fits on each frame for animate:true
            fit: true,
          
            // Padding in rendered co-ordinates around the layout
            padding: 30,
          
            // Constrain layout bounds with one of
            // - { x1, y1, x2, y2 }
            // - { x1, y1, w, h }
            // - undefined / null : Unconstrained
            boundingBox: undefined,
          
            // Layout event callbacks; equivalent to `layout.one('layoutready', callback)` for example
            ready: function(){}, // on layoutready
            stop: function(){}, // on layoutstop
          
            // Whether to randomize the initial positions of the nodes
            // true : Use random positions within the bounding box
            // false : Use the current node positions as the initial positions
            randomize: false
          };
        this.graph.layout(defaults).run();
    }

    getGraph() {
        this.masService.getMASById(this.selectedMASID.toString()).subscribe( (res: any) => {
            this.nodes = res.nodes.map(node => node.id);
            for (let item of res.nodes)
        })
    }

    drawGraph() {

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

    onClickList() {
        this.curr_state = "list";
    }

    onClickGraph() {
        this.curr_state = "graph";
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
            this.dfService.getAllSvcs(this.selectedMASID.toString()).subscribe( res => {
                this.searched_results = res;
                console.log(res);               
            },
            err => console.log(err)
            )
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
