package constants

import "time"

const LocalStorage = "localStorage/"
const ConvertedVideosBucketName = "converted_video_parts_1"
const SampleVideosBucketName = "videos_to_download"
const UnconvertedVideosBucketName = "uncoverted_video_parts_1"
const ProjectID = "fast-blueprint-296210"
const FinishedConversionExtension = ".converted"
const WorkManagementLoopSleepTime = time.Millisecond * 500
const DownloadChunkSizeInBytes = 1000
const SecondsBetweenVMCreationAndDeletion = 20
