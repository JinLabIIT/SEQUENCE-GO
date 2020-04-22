setwd("D:/QKD/SEQUENCE-GO-ringNetwork/R/mergeTime/data")

name = paste("k=", 2,".txt",sep="")
data <- read.csv(name,sep = " ",header = FALSE)
names(data) <- c("No.events","Time")
data$nlogk <- ceiling(log2(2))*data$No.events
data$k <- 2
for (i in 3:40){
  name = paste("k=", i,".txt",sep="")
  tmp <- read.csv(name,sep = " ",header = FALSE)
  names(tmp) <- c("No.events","Time")
  tmp$nlogk <- ceiling(log2(i))*tmp$No.events
  tmp$k <- i
  data <- rbind(data,tmp)
}

df1 <- data.frame()
for (i in seq(from = min(data$nlogk),to=max(data$nlogk),by= 1000)){
  tmp <- data[data$nlogk >i & data$nlogk < i+1000,]
  iqr = 1.5*IQR(tmp$Time)
  lower <- unname(quantile(tmp$Time,0.25))-iqr
  upper <- unname(quantile(tmp$Time,0.75))+iqr
  tmp <- tmp[tmp$Time > lower & tmp$Time < upper,]
  df1 <- rbind(df1,tmp)
}

df1_model2 <- lm(Time~No.events+nlogk,df1)

print(summary(df1_model2))