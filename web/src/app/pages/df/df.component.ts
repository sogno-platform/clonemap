import { Component, OnInit, Renderer2, Inject } from '@angular/core';
import { DfService} from "src/app/services/df.service";
import { MasService} from "src/app/services/mas.service";
import { ActivatedRoute, Params } from '@angular/router';
import * as cytoscape from 'cytoscape';
import popper from 'cytoscape-popper';
import { forkJoin, Observable } from 'rxjs';
import { Service } from 'src/app/models/service.model'
cytoscape.use( popper);


@Component({
  selector: 'app-df',
  templateUrl: './df.component.html',
  styleUrls: ['./df.component.css']
})
export class DFComponent implements OnInit {
    selectedMASID:number = -1;
    MASID: any = [];
    alive: boolean = true;
    fileToUpload: File = null;
    display: string = "";
    filename: string = "Choose a file...";
    searched_results: Service[] = [];
    curr_state: string = "list";
    graph: any;
    constructor(
        private dfService: DfService,
        private masService: MasService,
        private route: ActivatedRoute,
        ) { }

    ngOnInit() {
        this.dfService.getAlive().subscribe( (res: any) => {
            this.alive = res.df;
        }, error => {
            console.log(error);
        });

        // update the sidebar
        this.masService.getMAS().subscribe((MASs: any) => {
            if (MASs !== null) {
                for (let MAS of MASs) {
                    if (MAS.status.code != 5) {
                        this.MASID.push(MAS.id)
                    }
                }
            } 
        }, err => {
            console.log(err)  
        });

        this.route.params.subscribe((params: Params) => {
            if (params.masid) {
                this.selectedMASID = params.masid;
                this.dfService.getAllSvcs(this.selectedMASID.toString()).subscribe( (res: Service[]) => {
                    this.searched_results = res;  
                    for (let i = 0; i < this.searched_results.length; i++) {
                        this.searched_results[i].id = i;
                    }
                });       
            } else {
                console.log("No masid");
            }
        });
    }

    getNodeAndEdge(): Observable<any>  {
        return new Observable((observer) => {
            forkJoin({
                reqNode: this.masService.getMASById(this.selectedMASID.toString()),
                reqSvc: this.dfService.getAllSvcs(this.selectedMASID.toString())
            }).subscribe(({ reqNode, reqSvc } : any ) => {
                console.log(reqSvc)
                console.log(reqNode)
                const nodes = reqNode.graph.node.map(node => node.id);
                const agents = reqNode.agents.instances.map(agent => agent.id);
                const edgeNodes = reqNode.graph.edge;
                let edgeAgentNode = [];
                let svcs: string[] = [];
                let edgeSvcAgent = [];
                for (let i = 0; i < reqNode.graph.node.length; i++) {
                    for (let j = 0; j < reqNode.graph.node[i].agents.length; j++) {
                        edgeAgentNode.push({
                            n1: reqNode.graph.node[i].id,
                            n2: reqNode.graph.node[i].agents[j]
                        })
                    }
                }
                for (let i = 0; i < reqSvc.length; i++) {
                    svcs.push(reqSvc[i].desc);
                    edgeSvcAgent.push({
                        n1: i,
                        n2: reqSvc[i].agentid,
                    })
                }
                let res = {
                    "nodes": nodes,
                    "agents":agents,
                    "svcs": svcs,
                    "edgeNodes": edgeNodes,
                    "edgeAgentNode": edgeAgentNode,
                    "edgeSvcAgent": edgeSvcAgent
                }
                observer.next(res);
            })
        })

    }



    onClickList() {
        this.curr_state = "list";
    }

    onClickGraph() {
        this.curr_state = "graph";
        this.graph = cytoscape({
            container: document.getElementById('graph'),
            elements: [],
            zoom : 1,
            maxZoom: 2,
            minZoom: 0.5,
            zoomingEnabled: true,
            pan: {x: 500, y: 0},
            style: [
                {
                    selector: 'node',
                    style: {
                        'background-color': "#9696f3",
                        label: 'data(name)',
                        "text-halign": 'center',
                        "text-valign": 'center',
                    }
                },
                {
                    selector: ".node",
                    style: {
                        width: 120,
                        height: 120,
                        'background-color': "rgb(102,102,255)",
                        "font-size": 30,
                        color: 'white',
                }
            },
                {
                    selector: ".agent",
                    style: {
                        width: 90,
                        height: 90,
                        'background-color': "#9696f3",
                        "font-size": 20,
                        color: 'white',
                    }
                },
                {
                    selector: '.svc',
                    style: {
                        width: 70,
                        height: 70,
                        'background-color':"#dadaf8",
                        "font-size": 20,
                        color: 'grey',
                    }
                }
            ]
        });
        
        this.getNodeAndEdge().subscribe(res => {
            for (let i = 0; i < res.nodes.length; i++) {
                this.graph.add({
                    classes: "node",

                    data: { 
                        id: 'node' + res.nodes[i],
                        name: 'node' +  res.nodes[i].toString(),
                    }
                });            
            }

            for (let i = 0; i < res.agents.length; i++) {
                this.graph.add({
                    classes: 'agent', 
                    data: { 
                        id: 'agent' + res.agents[i],
                        name: 'agent' + res.agents[i].toString(),
                    }
                });
            }

            for (let i = 0; i < res.svcs.length; i++) {
                this.graph.add({
                    classes: 'svc',
                    data: { 
                        id: 'svc' + i,
                        name: res.svcs[i],
                        }
                });
            }

            let k = 0;
            // edges between the nodes
            for (let i = 0; i < res.edgeNodes.length; i++) {
                this.graph.add({
                    data: {
                        id: 'edge' + k,
                        source: 'node' + res.edgeNodes[i].n1,
                        target: 'node' + res.edgeNodes[i].n2
                    }
                })
                k++;
            }

            // edges between agent and nodes
            for (let i = 0; i < res.edgeAgentNode.length; i++) {
                    this.graph.add({
                        data: {
                            id: 'edge' + k,
                            source: 'node' + res.edgeAgentNode[i].n1,
                            target: 'agent' + res.edgeAgentNode[i].n2
                        }
                    })
                    k++;
            }

            // edges between services and agent
            for (let i = 0; i < res.edgeSvcAgent.length; i++) {
                    this.graph.add({
                        data: {
                            id: 'edge' + k,
                            source: 'svc' + res.edgeSvcAgent[i].n1,
                            target: 'agent' + res.edgeSvcAgent[i].n2
                        }
                    })
                    k++;
            }
            this.graph.layout({
                name : "cose",
                fit: true,
            }).run();   

            this.graph.$('node').on('tap', function(evt){
                console.log( 'tap ' + evt.target.id() );
            });
        })  
    }


    onSearchSvcs(desc:string, nodeid:string, dist:string) {
        let masid: string = this.selectedMASID.toString();
    
        if (desc === "" && nodeid === "" && dist === "") {
            this.dfService.getAllSvcs(masid).subscribe( (res: Service[]) => {
                this.searched_results = res;   
                for (let i = 0; i < this.searched_results.length; i++) {
                    this.searched_results[i].id = i;
                }           
            },
            err => console.log(err)
            )
        }

        else if (desc !== "" && nodeid == "" && dist == "") {
            this.dfService.searchSvc(masid, desc).subscribe( (res: Service[])  => {
                this.searched_results = res; 
                for (let i = 0; i < this.searched_results.length; i++) {
                    this.searched_results[i].id = i;
                }
            },
            err => console.log(err)
            )
        }

        else if (desc !== "" && nodeid !== "" && dist !== "") {
            this.dfService.searchSvcWithinDis(masid, desc, nodeid, dist).subscribe( 
                (res: Service[])  => {
                    this.searched_results = res;
                    for (let i = 0; i < this.searched_results.length; i++) {
                        this.searched_results[i].id = i;
                    }
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
