export class MAS {
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

   imagegroups: Array<{
          config:{
              image: string
          },
          agents:Array<{
                  nodeid: number,
                  name: string,
                  type: string,
                  custom: string
              }>
          
      }>;

   graph: {
      node: string,
      edge: string
   };


}

