package controller

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type S3 struct {
	controller.Abstract
}

type S3ConnTestRequest struct {
	AppID  string `json:"appid" binding:"required"`
	Secret string `json:"secret" binding:"required"`
	Region string `json:"region" binding:"required"`
	Bucket string `json:"bucket" binding:"required"`
}

func (self *S3) ConnTest(c *gin.Context) {
	var req S3ConnTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if !self.Validate(c, &req) {
		return
	}

	// 创建AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(req.Region),
		Credentials: credentials.NewStaticCredentials(req.AppID, req.Secret, ""),
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to create AWS session"})
		return
	}

	// 创建S3客户端
	svc := s3.New(sess)

	// 测试列出bucket
	_, err = svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(req.Bucket),
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to connect to S3: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "S3 connection successful"})
}
