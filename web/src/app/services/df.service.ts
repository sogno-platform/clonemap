import { Injectable } from '@angular/core';
import { WebRequestService } from './web-request.service';

@Injectable({
  providedIn: 'root'
})
export class DfService {

  constructor(private webReqService: WebRequestService) { }
  
  getAlive() {
    return this.webReqService.getText('alive/df');
  }
  
  getAllSvcs(masid: number) {
    return this.webReqService.get(`api/df/${masid}/svc`);
  }
  
  createSvc(masid: number, payload: object) {
    return this.webReqService.post(`api/df/${masid}/svc`, payload);
  }

  getGraph(masid: number) {
    return this.webReqService.get(`api/df/${masid}/graph`);
  }

  createGraph(masid: number, payload: object) {
    return this.webReqService.post(`api/df/${masid}/graph`, payload);
  }

  updateGraph(masid: number, payload: object) {
    return this.webReqService.patch(`api/df/${masid}/graph`, payload);
  }

  searchSvc(masid: number, desc: string) {
    return this.webReqService.get(`api/df/${masid}/svc/desc/${desc}`);
  }

  searchSvcWithinDis(masid: number, desc: string, nodeid: number, dist: number) {
    return this.webReqService.get(`api/df/${masid}/svc/desc/${desc}/node/${nodeid}/dist/${dist}`);
  }

  searchSvcById(masid: number, svcid: number) {
    return this.webReqService.get(`api/df/${masid}/svc/id/${svcid}`);
  }
  
  deleteSvcById(masid:number, svcid: number) {
    return this.webReqService.delete(`api/df/${masid}/svc/id/${svcid}`);
  }
  
  
}
