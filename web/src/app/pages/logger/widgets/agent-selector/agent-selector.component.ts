import { Component, OnInit, Input, Output } from '@angular/core';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { DefaultAMSService } from 'src/app/openapi-services/ams';
import { EventEmitter } from 'stream';

@Component({
  selector: 'app-agent-selector',
  templateUrl: './agent-selector.component.html',
  styleUrls: ['./agent-selector.component.scss']
})
export class AgentSelectorComponent implements OnInit {
    @Input() selectedMASID: string = "1";
    agentID: number[];
    selectedID: number[] = [];
    notSelectedID: number[] = [];
    isAgentSelected: boolean[] = [];
      
    constructor(
        private modalService: NgbModal,
        private amsService: DefaultAMSService,
    ) { }


    ngOnInit(): void {
    
            this.amsService.getOneMAS(parseInt(this.selectedMASID)).subscribe((res: any) => {
                if (res.agents.counter !== 0) {
                    this.agentID = res.agents.instances.map(item => item.id);
                    for (let i = 0; i < res.agents.counter; i++) {
                        this.isAgentSelected.push(false);
                    }
                    this.updateSelectedID();
                }
            });
    }

    onDeleteID(i : number) {
        this.isAgentSelected[i] = !this.isAgentSelected[i];
        this.updateSelectedID();
    }

    openLg(content) {
        this.modalService.open(content, { size: 'lg', centered: true });
    }

    onAddID(i: number) {
        if (this.selectedID.length < 10) {
            this.isAgentSelected[i] = !this.isAgentSelected[i];
            this.updateSelectedID();

        }
    }

    updateSelectedID() {
        this.selectedID = [];
        this.notSelectedID = [];
        for (let i = 0; i < this.agentID.length; i++) {
            if (this.isAgentSelected[i]) {
                this.selectedID.push(i);
            } else {
                this.notSelectedID.push(i);
            }
        }
    }


    onConfirm() {
        this.modalService.dismissAll();       
    }
}
