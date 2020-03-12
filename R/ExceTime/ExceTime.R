library(MLmetrics)
library(MASS)
setwd("/Users/cheung/Desktop/tmp/exceTime")
set.seed(60616)
df <- read.csv("ExceTime.txt",sep = ",",header = FALSE)
names(df) <- c("exceTime","No.events","k")
#df <- df[df$No.events!=0, ]
df_lm <- lm(exceTime~.,df) 

df <- df[df$k<=32,]
df <- df[df$No.events!=0, ]
#df <- df[df$No.events < 13000,]
#df$ksquare <- df$k*df$k*df$No.events
#df$kcube <- df$k*df$k*df$k*df$No.events
df$nk <- df$No.events*df$k
df_lm2 <- lm(exceTime~No.events,df)

train <- sample(1:nrow(df),.8*nrow(df))
xtest <-df[-train,c(2,4)]
ytest <- df[-train,1]
xtrain <-df[train,]
ytrain <- df[train,1]
model <- lm(exceTime~No.events+nk,df[train,])
result <- predict(model,xtest,interval = "prediction")
fit <- result[,1]

SSR = sum((fit-ytest)^2)
SST = sum((fit-mean(ytest))^2)
r_square = 1 - SSR/SST

aaa <- df[-train,]
mean(abs(ytest[aaa$k == 2] - fit[aaa$k == 2] )/ytest[aaa$k == 2])
mean(abs(ytest[aaa$k == 4] - fit[aaa$k == 4] )/ytest[aaa$k == 4])
mean(abs(ytest[aaa$k == 6] - fit[aaa$k == 6] )/ytest[aaa$k == 6])
mape1 <- c()
for (m in seq(from = 2, to = 32, by = 2)){
  mape1 <- c(mape1, mean(abs(ytest[aaa$k == m] - fit[aaa$k == m] )/ytest[aaa$k == m]))
}
mean(abs(ytest[aaa$k == m] - fit[aaa$k == m] )/ytest[aaa$k == m])
#mean(abs(ytest[10000:165594] - fit[10000:165594] )/ytest[10000:165594])

#boxcox(lm(exceTime~nk,df[train,]),lambda=seq(0,1,by=.1))