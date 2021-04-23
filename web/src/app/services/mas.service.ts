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

    getclonemap() {
        return this.webReqService.get('api/clonemap');
    }

    getMAS() {
        return this.webReqService.get('api/clonemap/mas');
    }

    createMAS(payload: object) {
        return this.webReqService.post('api/clonemap/mas', payload);
    }

    getMASById(masid: string) {
        return this.webReqService.get(`api/clonemap/mas/${masid}`);
    }

    deleteMASById(masid: string) {
        return this.webReqService.delete(`api/clonemap/mas/${masid}`);
    }

    getAgents(masid: string) {
        return this.webReqService.get(`api/clonemap/mas/${masid}/agents`);
    }

    addAgents(masid: string, payload: object) {
        return this.webReqService.post(`api/clonemap/mas/${masid}/agents`, payload );
    }

    getAgentById(masid:string, agentid: string) {
        return this.webReqService.get(`api/clonemap/mas/${masid}/agents/${agentid}`);
    }

    deleteAgentById(masid:string, agentid: string) {
        return this.webReqService.delete(`api/clonemap/mas/${masid}/agents/${agentid}`);
    }

    getAgentAdress(masid:string, agentid: string) {
        return this.webReqService.get(`api/clonemap/mas/${masid}/agents/${agentid}/address`);
    }

    updateAgentAddress(masid: string, agentid:string, payload: object) {
        return this.webReqService.post(`api/clonemap/mas/${masid}/agents/${agentid}/address`, payload);
    }

    customUpdateAgentAddress(masid: string, agentid: string, payload: object) {
        return this.webReqService.post(`api/clonemap/mas/${masid}/agents/${agentid}/address`, payload);
        
    }

    getAgentbyName(masid:string, name: string) {
        return this.webReqService.get(`api/clonemap/mas/${masid}/agents/name/${name}`);

    }

    getAllAgencies(masid: string) {
        return this.webReqService.get(`api/clonemap/mas/${masid}/agencies`);
    }

    getAgencyInformation(masid: string, imid: string, agencyid: string) {
        return this.webReqService.get(`api/clonemap/mas/${masid}/imgroup/${imid}/agencies/${agencyid}`);
    }



}



