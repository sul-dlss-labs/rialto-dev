FROM openjdk:8-jre

ADD blazegraph.jar blazegraph.jar
ADD BGStore.properties BGStore.properties

EXPOSE 9999

CMD java -server -Xmx4g -Dcom.bigdata.rdf.sail.webapp.ConfigParams.propertyFile=BGStore.properties -jar blazegraph.jar
