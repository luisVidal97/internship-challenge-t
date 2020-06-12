This is a back-end of intership challenge.
The structure of this app  is the following:
- conndb.go file contain the code for create the connetion with database  (CockroachDB) and
queries in this database used in the main method in main.go file
- getLogoTitle.go file contain the logic to get logo and title through the request of damin
and checking  the labels in html code.
- main.go file is the file where is endpoint method( main() ) that execute all program. Here
are routes with fasthttprouter and the methods to get information Of WHOIS and SSLlabs Apis. 