import { Injectable } from '@angular/core';
import { WebRequestService } from './web-request.service'

@Injectable({
  providedIn: 'root'
})
export class MasService {

  constructor(private webReqService: WebRequestService) { }

    getAlive() {
        return this.webReqService.get('api/alive');
    }

    getCloneMAP() {
        return this.webReqService.get('api/clonemap');
    }

    getMAS() {
        return this.webReqService.get('api/clonemap/mas');
    }

    createMAP(payload: object) {
        return this.webReqService.post('api/clonemap/mas', payload);
    }

    getMASById(masid: number) {
        return this.webReqService.get(`api/clonemap/mas/${masid}`);
    }

    deleteMASById(masid: number) {
        return this.webReqService.delete(`api/clonemap/mas/${masid}`);
    }

    getAgents(masid: number) {
        return this.webReqService.get(`api/clonemap/mas/${masid}/agents`);
    }

    addAgents(masid: number, payload: object) {
        return this.webReqService.post(`api/clonemap/mas/${masid}/agents`, payload );
    }

    getAgentById(masid:number, agentid: number) {
        return this.webReqService.get(`api/clonemap/mas/${masid}/agents/${agentid}`);
    }

    deleteAgentById(masid:number, agentid: number) {
        return this.webReqService.delete(`api/clonemap/mas/${masid}/agents/${agentid}`);
    }

    getAgentAdress(masid:number, agentid: number) {
        return this.webReqService.get(`api/clonemap/mas/${masid}/agents/${agentid}/address`);
    }

    updateAgentAddress(masid: Number, agentid:number, payload: object) {
        return this.webReqService.post(`api/clonemap/mas/${masid}/agents/${agentid}/address`, payload);
    }

    customUpdateAgentAddress(masid: Number, agentid: number, payload: object) {
        return this.webReqService.post(`api/clonemap/mas/${masid}/agents/${agentid}/address`, payload);
        
    }

    getAgentbyName(masid:number, name: string) {
        return this.webReqService.get(`api/clonemap/mas/${masid}/agents/name/${name}`);

    }

    getAllAgencies(masid: number) {
        return this.webReqService.get(`api/clonemap/mas/${masid}/agencies`);
    }

    getAgencyInformation(masid: number, imid: number, agencyid: number) {
        return this.webReqService.get(`api/clonemap/mas/${masid}/imgroup/${imid}/agencies/${agencyid}`);
    }



}



