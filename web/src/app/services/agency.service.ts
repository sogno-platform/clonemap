import { Injectable } from '@angular/core';
import { WebRequestService} from './web-request.service'

@Injectable({
  providedIn: 'root'
})
export class AgencyService {

  constructor(private webReqService: WebRequestService) { }
  
  getAgency() {
    return this.webReqService.get('/api/agency');
  }

  createAgent(payload: object) {
    return this.webReqService.post('./api/agency/agents', payload);
  }

  createMsgs(payload: object) {
    return this.webReqService.post('/api/agency/msgs', payload);
  }

  createMsgsUndeliv(payload: object) {
    return this.webReqService.post('/api/agency/msgundeliv', payload);
  }

  deleteAgent(agentid: number) {
    return this.webReqService.delete(`/api/agency/agents/${agentid}`);
  }

  getAgentStatus(agentid: number) {
    return this.webReqService.get(`/api/agency/agents/${agentid}/status`);
  }

  updateAgentStatus(agentid: number, payload: object) {
    return this.webReqService.patch(`/api/agency/agents/${agentid}/custom`, payload);
  }
}
