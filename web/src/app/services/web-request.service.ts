import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http'

@Injectable({
  providedIn: 'root'
})
export class WebRequestService {
    readonly ROOT_URL;
    readonly options;
    readonly headerDict = {
        'Content-Type': 'application/json',
    };

    constructor(private http: HttpClient) {
        this.ROOT_URL = 'http://localhost:4200';
          //this.contents = '';
        this.options = {                                                                                                                                                                                 
            headers: new HttpHeaders(this.headerDict), 
        };
    }



    get(uri: string) {
        return this.http.get(`${this.ROOT_URL}/${uri}`);
    }

    post(uri: string, payload: object) {
        return this.http.post<any>(`${this.ROOT_URL}/${uri}`, payload, this.options);
    }

    patch(uri: string, payload: Object) {
        return this.http.patch(`${this.ROOT_URL}/${uri}`, payload);
    }

    delete(uri: string) {
        return this.http.delete(`${this.ROOT_URL}/${uri}`);
    }

}
