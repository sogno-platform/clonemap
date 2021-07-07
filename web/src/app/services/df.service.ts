import { Injectable } from '@angular/core';
import { WebRequestService } from './web-request.service';

@Injectable({
  providedIn: 'root'
})
export class DfService {

  constructor(private webReqService: WebRequestService) { }
  
  getAlive() {
    return this.webReqService.get('api/pf/modules');
  }
  
  getAllSvcs(masid: string) {
    return this.webReqService.get(`api/df/${masid}/svc`);
  }
  
  createSvc(masid: string, payload: object) {
    return this.webReqService.post(`api/df/${masid}/svc`, payload);
  }

  getGraph(masid: string) {
    return this.webReqService.get(`api/df/${masid}/graph`);
  }

  createGraph(masid: string, payload: object) {
    return this.webReqService.post(`api/df/${masid}/graph`, payload);
  }

  updateGraph(masid: string, payload: object) {
    return this.webReqService.patch(`api/df/${masid}/graph`, payload);
  }

  searchSvc(masid: string, desc: string) {
    return this.webReqService.get(`api/df/${masid}/svc/desc/${desc}`);
  }

  searchSvcWithinDis(masid: string, desc: string, nodeid: string, dist: string) {
    return this.webReqService.get(`api/df/${masid}/svc/desc/${desc}/node/${nodeid}/dist/${dist}`);
  }

  searchSvcById(masid: string, svcid: string) {
    return this.webReqService.get(`api/df/${masid}/svc/id/${svcid}`);
  }
  
  
}
