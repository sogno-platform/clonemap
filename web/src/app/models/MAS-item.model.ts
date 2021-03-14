export class MASItem {
   id: number;
   config: {
       name: string,
       agentsperagency: number,
       mqtt: {
           active: boolean
       },
       df: {
           active: boolean
       },
       logger: {
           active: boolean,
           msg: boolean,
           app: boolean,
           status: boolean,
           debug: boolean
       }
   };
   graph: {
       node: Array <{
               id: number,
               agents: number[]
           }>,
       edge: null
   };

   imagegroups: {
       counter: number,
       instances: Array<{
               config: {
                   image: string
               },
               id: number,
               agencies: {
                   counter: number,
                   instances: Array<{
                        masid: number,
                        name: string,
                        id: number,
                        imid: number,
                        logger: {
                              active: boolean,
                              msg: boolean,
                              app: boolean,
                              status: boolean,
                              debug: boolean
                        },
                        agents: number[],
                        status: {
                              code: number,
                              lastupdate: Date
                        }
                  }>
               }
           }>
   };

   agents: {
       counter: number,
       instances: Array<{
               spec: {
                   nodeid: number,
                   name: string,
                   type: string
               },
               masid: number,
               agencyid: number,
               imid: number,
               id: number,
               address: {
                   agency: string
               },
               status: {
                   code: number,
                   lastupdate: Date
               }
           }>
   };

   uptime: Date;

   status: {
       code: number,
       lastupdate: Date
   };
}