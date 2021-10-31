openapi-generator-cli generate -g typescript-angular -i ./agency/openapi.yaml -o ../../web/src/app/openapi-services/agency --additional-properties npmName=agency-api,snapshot=false,ngVersion=12.2.9,npmVersion=6.14.14,apiModulePrefix=Agency,serviceSuffix=AgencyService

openapi-generator-cli generate -g typescript-angular -i ./ams/openapi.yaml -o ../../web/src/app/openapi-services/ams --additional-properties npmName=ams-api,snapshot=false,ngVersion=12.2.9,npmVersion=6.14.14,apiModulePrefix=AMS,serviceSuffix=AMSService

openapi-generator-cli generate -g typescript-angular -i ./df/openapi.yaml -o ../../web/src/app/openapi-services/df --additional-properties npmName=df-api,snapshot=false,ngVersion=12.2.9,npmVersion=6.14.14,apiModulePrefix=DF,serviceSuffix=DFService

openapi-generator-cli generate -g typescript-angular -i ./logger/openapi.yaml -o ../../web/src/app/openapi-services/logger --additional-properties npmName=logger-api,snapshot=false,ngVersion=12.2.9,npmVersion=6.14.14,apiModulePrefix=Logger,serviceSuffix=LoggerService


