import { Injectable } from '@angular/core';
import { WebRequestService} from './web-request.service';

@Injectable({
  providedIn: 'root'
})
export class LoggerService {

  constructor(private webReqService: WebRequestService) { }

  getAlive() {
    return this.webReqService.get('api/pf/modules')
  }

  getNLatestLogs(masid: string, agentid: string, topic:string, num: string) {
    return this.webReqService.get(`api/logging/${masid}/${agentid}/${topic}/latest/${num}`);
  }
  
  getLogsInRange(masid: string, agentid: string, topic: string, start: string, end: string) {
    return this.webReqService.get(`api/logging/${masid}/${agentid}/${topic}/time/${start}/${end}`);
  }

  getLogSeriesNames(masid:string, agentid: string) {
    return this.webReqService.get(`api/logging/series/${masid}/${agentid}/names`)
  }

  getLogSeriesByName(masid:string, agentid: string, name: string, start: string, end: string) {
    return this.webReqService.get(`api/logging/series/${masid}/${agentid}/${name}/time/${start}/${end}`)
  }

  getMsgHeatmap(masid: string, start: string, end: string) {
    return this.webReqService.get(`api/logging/stats/${masid}/heatmap/${start}/${end}`)
  }

  getBehavior(masid:string, agentid: string, behtype: string, start: string, end: string) {
    return this.webReqService.get(`api/logging/stats/${masid}/${agentid}/${behtype}/${start}/${end}`)
  }
}
