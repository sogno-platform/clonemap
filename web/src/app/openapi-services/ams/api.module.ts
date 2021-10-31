import { NgModule, ModuleWithProviders, SkipSelf, Optional } from '@angular/core';
import { Configuration } from './configuration';
import { HttpClient } from '@angular/common/http';

import { DefaultAMSService } from './api/default.service';

@NgModule({
  imports:      [],
  declarations: [],
  exports:      [],
  providers: []
})
export class AMSApiModule {
    public static forRoot(configurationFactory: () => Configuration): ModuleWithProviders<AMSApiModule> {
        return {
            ngModule: AMSApiModule,
            providers: [ { provide: Configuration, useFactory: configurationFactory } ]
        };
    }

    constructor( @Optional() @SkipSelf() parentModule: AMSApiModule,
                 @Optional() http: HttpClient) {
        if (parentModule) {
            throw new Error('AMSApiModule is already loaded. Import in your base AppModule only.');
        }
        if (!http) {
            throw new Error('You need to import the HttpClientModule in your AppModule! \n' +
            'See also https://github.com/angular/angular/issues/20575');
        }
    }
}
