package benchmark.DF;

import java.util.Random;

import jade.core.*;
import jade.core.behaviours.*;
import jade.lang.acl.ACLMessage;
import jade.domain.FIPAAgentManagement.ServiceDescription;
import jade.domain.FIPAAgentManagement.DFAgentDescription;
import jade.domain.DFService;
import jade.domain.FIPAException;
import jade.util.Logger;

/**
 */
public class DF extends Agent {

	private Logger myLogger = Logger.getMyLogger(getClass().getName());

	private class DFBehaviour extends OneShotBehaviour{
        Agent agnt;
        long[] rtts;

		public DFBehaviour(Agent ag) {
            super(ag);
            agnt = ag;
            rtts = new long[11];
            Object[] args = getArguments();
            try{
                Thread.sleep(10000);
            }
            catch (InterruptedException temp){
            }
		}

		public void action() {
            Random r = new Random();
            for (int i = 0; i < 100; i++) {
                DFAgentDescription searchAgentTemplate = new DFAgentDescription();
                ServiceDescription searchServiceTemplate = new ServiceDescription();
                searchServiceTemplate.setType("svc"+String.valueOf(r.nextInt(7)));
                searchAgentTemplate.addServices(searchServiceTemplate);
                try{
                    DFService.search(agnt, searchAgentTemplate);
                }
                catch (FIPAException fe) {
                    fe.printStackTrace();
                }
            }

            long startTime = System.nanoTime();

            try {
                DFAgentDescription registerAgentTemplate = new DFAgentDescription();
                registerAgentTemplate.setName(getAID());
                ServiceDescription registerServiceTemplate = new ServiceDescription();
                registerServiceTemplate.setName(getAID().getName());
                registerServiceTemplate.setType("svc"+String.valueOf(r.nextInt(7)));
                registerAgentTemplate.addServices(registerServiceTemplate);
                
                DFService.register(agnt, registerAgentTemplate);
            }
            catch (FIPAException fe) {
                fe.printStackTrace();
            }

            long regTime = System.nanoTime();
            rtts[0] = (regTime - startTime) / 1000;

            for (int i = 0; i < 8; i++) {
                long searchStart = System.nanoTime();
                DFAgentDescription searchAgentTemplate = new DFAgentDescription();
                ServiceDescription searchServiceTemplate = new ServiceDescription();
                searchServiceTemplate.setType("svc"+String.valueOf(r.nextInt(7)));
                searchAgentTemplate.addServices(searchServiceTemplate);
                try{
                    DFService.search(agnt, searchAgentTemplate);
                }
                catch (FIPAException fe) {
                    fe.printStackTrace();
                }
                long searchStop = System.nanoTime();
                rtts[i+1] = (searchStop-searchStart) / 1000;
            }

            long deregStart = System.nanoTime();

            try{
                DFService.deregister(agnt);
            }
            catch(FIPAException fe) {
                fe.printStackTrace();
            }

            long stopTime = System.nanoTime();

            rtts[9] = (stopTime - deregStart) / 1000;
            rtts[10] = (stopTime - startTime)/1000; 

            for (int i = 0; i < 100; i++) {
                DFAgentDescription searchAgentTemplate = new DFAgentDescription();
                ServiceDescription searchServiceTemplate = new ServiceDescription();
                searchServiceTemplate.setType("svc"+String.valueOf(r.nextInt(7)));
                searchAgentTemplate.addServices(searchServiceTemplate);
                try{
                    DFService.search(agnt, searchAgentTemplate);
                }
                catch (FIPAException fe) {

                }
            }

            String content = "";
            for (int i = 0; i < 11; i++) {
                content += String.valueOf(rtts[i]);
                if (i < 10) {
                    content+= ";";
                }
            }
            ACLMessage logmsg = new ACLMessage(ACLMessage.INFORM);
            logmsg.setSender(myAgent.getAID());
            AID receiver = new AID("logag", AID.ISLOCALNAME);
            logmsg.addReceiver(receiver);
            logmsg.setProtocol("p");

            logmsg.setContent(content);
            send(logmsg);
        }
        
    }

	protected void setup() {
		DFBehaviour df = new  DFBehaviour(this);
        addBehaviour(df);
	}
}
