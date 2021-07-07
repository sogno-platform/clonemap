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

    getMAS() {
        return this.webReqService.get('api/ams/mas');
    }

    createMAS(payload: object) {
        return this.webReqService.post('api/ams/mas', payload);
    }

    getMASById(masid: string) {
        return this.webReqService.get(`api/ams/mas/${masid}`);
    }

    deleteMASById(masid: string) {
        return this.webReqService.delete(`api/ams/mas/${masid}`);
    }
}



