libvenice.so: venice.o
	gcc -fPIC -shared -o libvenice.so venice.o

venice.o: venice.h venice.c
	gcc -fPIC -c -o venice.o venice.c

clean:
	rm -f *.o *.so
