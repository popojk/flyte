# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc

from flyteidl.datacatalog import datacatalog_pb2 as flyteidl_dot_datacatalog_dot_datacatalog__pb2


class DataCatalogStub(object):
    """
    Data Catalog service definition
    Data Catalog is a service for indexing parameterized, strongly-typed data artifacts across revisions.
    Artifacts are associated with a Dataset, and can be tagged for retrieval.
    """

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.CreateDataset = channel.unary_unary(
                '/datacatalog.DataCatalog/CreateDataset',
                request_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.CreateDatasetRequest.SerializeToString,
                response_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.CreateDatasetResponse.FromString,
                )
        self.GetDataset = channel.unary_unary(
                '/datacatalog.DataCatalog/GetDataset',
                request_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetDatasetRequest.SerializeToString,
                response_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetDatasetResponse.FromString,
                )
        self.CreateArtifact = channel.unary_unary(
                '/datacatalog.DataCatalog/CreateArtifact',
                request_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.CreateArtifactRequest.SerializeToString,
                response_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.CreateArtifactResponse.FromString,
                )
        self.GetArtifact = channel.unary_unary(
                '/datacatalog.DataCatalog/GetArtifact',
                request_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetArtifactRequest.SerializeToString,
                response_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetArtifactResponse.FromString,
                )
        self.CreateFutureArtifact = channel.unary_unary(
                '/datacatalog.DataCatalog/CreateFutureArtifact',
                request_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.CreateArtifactRequest.SerializeToString,
                response_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.CreateArtifactResponse.FromString,
                )
        self.GetFutureArtifact = channel.unary_unary(
                '/datacatalog.DataCatalog/GetFutureArtifact',
                request_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetArtifactRequest.SerializeToString,
                response_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetArtifactResponse.FromString,
                )
        self.UpdateFutureArtifact = channel.unary_unary(
                '/datacatalog.DataCatalog/UpdateFutureArtifact',
                request_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.UpdateArtifactRequest.SerializeToString,
                response_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.UpdateArtifactResponse.FromString,
                )
        self.AddTag = channel.unary_unary(
                '/datacatalog.DataCatalog/AddTag',
                request_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.AddTagRequest.SerializeToString,
                response_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.AddTagResponse.FromString,
                )
        self.ListArtifacts = channel.unary_unary(
                '/datacatalog.DataCatalog/ListArtifacts',
                request_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.ListArtifactsRequest.SerializeToString,
                response_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.ListArtifactsResponse.FromString,
                )
        self.ListDatasets = channel.unary_unary(
                '/datacatalog.DataCatalog/ListDatasets',
                request_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.ListDatasetsRequest.SerializeToString,
                response_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.ListDatasetsResponse.FromString,
                )
        self.UpdateArtifact = channel.unary_unary(
                '/datacatalog.DataCatalog/UpdateArtifact',
                request_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.UpdateArtifactRequest.SerializeToString,
                response_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.UpdateArtifactResponse.FromString,
                )
        self.GetOrExtendReservation = channel.unary_unary(
                '/datacatalog.DataCatalog/GetOrExtendReservation',
                request_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetOrExtendReservationRequest.SerializeToString,
                response_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetOrExtendReservationResponse.FromString,
                )
        self.ReleaseReservation = channel.unary_unary(
                '/datacatalog.DataCatalog/ReleaseReservation',
                request_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.ReleaseReservationRequest.SerializeToString,
                response_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.ReleaseReservationResponse.FromString,
                )


class DataCatalogServicer(object):
    """
    Data Catalog service definition
    Data Catalog is a service for indexing parameterized, strongly-typed data artifacts across revisions.
    Artifacts are associated with a Dataset, and can be tagged for retrieval.
    """

    def CreateDataset(self, request, context):
        """Create a new Dataset. Datasets are unique based on the DatasetID. Datasets are logical groupings of artifacts.
        Each dataset can have one or more artifacts
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetDataset(self, request, context):
        """Get a Dataset by the DatasetID. This returns the Dataset with the associated metadata.
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def CreateArtifact(self, request, context):
        """Create an artifact and the artifact data associated with it. An artifact can be a hive partition or arbitrary
        files or data values
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetArtifact(self, request, context):
        """Retrieve an artifact by an identifying handle. This returns an artifact along with the artifact data.
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def CreateFutureArtifact(self, request, context):
        """Create future artifact data.
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetFutureArtifact(self, request, context):
        """Retrieve a future artifact by an identifying handle. This returns an artifact along with the artifact data.
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def UpdateFutureArtifact(self, request, context):
        """Updates an existing future artifact, overwriting the stored artifact data in the underlying blob storage.
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def AddTag(self, request, context):
        """Associate a tag with an artifact. Tags are unique within a Dataset.
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def ListArtifacts(self, request, context):
        """Return a paginated list of artifacts
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def ListDatasets(self, request, context):
        """Return a paginated list of datasets
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def UpdateArtifact(self, request, context):
        """Updates an existing artifact, overwriting the stored artifact data in the underlying blob storage.
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetOrExtendReservation(self, request, context):
        """Attempts to get or extend a reservation for the corresponding artifact. If one already exists
        (ie. another entity owns the reservation) then that reservation is retrieved.
        Once you acquire a reservation, you need to  periodically extend the reservation with an
        identical call. If the reservation is not extended before the defined expiration, it may be
        acquired by another task.
        Note: We may have multiple concurrent tasks with the same signature and the same input that
        try to populate the same artifact at the same time. Thus with reservation, only one task can
        run at a time, until the reservation expires.
        Note: If task A does not extend the reservation in time and the reservation expires, another
        task B may take over the reservation, resulting in two tasks A and B running in parallel. So
        a third task C may get the Artifact from A or B, whichever writes last.
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def ReleaseReservation(self, request, context):
        """Release the reservation when the task holding the spot fails so that the other tasks
        can grab the spot.
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_DataCatalogServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'CreateDataset': grpc.unary_unary_rpc_method_handler(
                    servicer.CreateDataset,
                    request_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.CreateDatasetRequest.FromString,
                    response_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.CreateDatasetResponse.SerializeToString,
            ),
            'GetDataset': grpc.unary_unary_rpc_method_handler(
                    servicer.GetDataset,
                    request_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetDatasetRequest.FromString,
                    response_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetDatasetResponse.SerializeToString,
            ),
            'CreateArtifact': grpc.unary_unary_rpc_method_handler(
                    servicer.CreateArtifact,
                    request_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.CreateArtifactRequest.FromString,
                    response_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.CreateArtifactResponse.SerializeToString,
            ),
            'GetArtifact': grpc.unary_unary_rpc_method_handler(
                    servicer.GetArtifact,
                    request_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetArtifactRequest.FromString,
                    response_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetArtifactResponse.SerializeToString,
            ),
            'CreateFutureArtifact': grpc.unary_unary_rpc_method_handler(
                    servicer.CreateFutureArtifact,
                    request_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.CreateArtifactRequest.FromString,
                    response_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.CreateArtifactResponse.SerializeToString,
            ),
            'GetFutureArtifact': grpc.unary_unary_rpc_method_handler(
                    servicer.GetFutureArtifact,
                    request_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetArtifactRequest.FromString,
                    response_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetArtifactResponse.SerializeToString,
            ),
            'UpdateFutureArtifact': grpc.unary_unary_rpc_method_handler(
                    servicer.UpdateFutureArtifact,
                    request_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.UpdateArtifactRequest.FromString,
                    response_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.UpdateArtifactResponse.SerializeToString,
            ),
            'AddTag': grpc.unary_unary_rpc_method_handler(
                    servicer.AddTag,
                    request_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.AddTagRequest.FromString,
                    response_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.AddTagResponse.SerializeToString,
            ),
            'ListArtifacts': grpc.unary_unary_rpc_method_handler(
                    servicer.ListArtifacts,
                    request_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.ListArtifactsRequest.FromString,
                    response_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.ListArtifactsResponse.SerializeToString,
            ),
            'ListDatasets': grpc.unary_unary_rpc_method_handler(
                    servicer.ListDatasets,
                    request_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.ListDatasetsRequest.FromString,
                    response_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.ListDatasetsResponse.SerializeToString,
            ),
            'UpdateArtifact': grpc.unary_unary_rpc_method_handler(
                    servicer.UpdateArtifact,
                    request_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.UpdateArtifactRequest.FromString,
                    response_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.UpdateArtifactResponse.SerializeToString,
            ),
            'GetOrExtendReservation': grpc.unary_unary_rpc_method_handler(
                    servicer.GetOrExtendReservation,
                    request_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetOrExtendReservationRequest.FromString,
                    response_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetOrExtendReservationResponse.SerializeToString,
            ),
            'ReleaseReservation': grpc.unary_unary_rpc_method_handler(
                    servicer.ReleaseReservation,
                    request_deserializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.ReleaseReservationRequest.FromString,
                    response_serializer=flyteidl_dot_datacatalog_dot_datacatalog__pb2.ReleaseReservationResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'datacatalog.DataCatalog', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class DataCatalog(object):
    """
    Data Catalog service definition
    Data Catalog is a service for indexing parameterized, strongly-typed data artifacts across revisions.
    Artifacts are associated with a Dataset, and can be tagged for retrieval.
    """

    @staticmethod
    def CreateDataset(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/datacatalog.DataCatalog/CreateDataset',
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.CreateDatasetRequest.SerializeToString,
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.CreateDatasetResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def GetDataset(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/datacatalog.DataCatalog/GetDataset',
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetDatasetRequest.SerializeToString,
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetDatasetResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def CreateArtifact(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/datacatalog.DataCatalog/CreateArtifact',
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.CreateArtifactRequest.SerializeToString,
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.CreateArtifactResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def GetArtifact(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/datacatalog.DataCatalog/GetArtifact',
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetArtifactRequest.SerializeToString,
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetArtifactResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def CreateFutureArtifact(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/datacatalog.DataCatalog/CreateFutureArtifact',
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.CreateArtifactRequest.SerializeToString,
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.CreateArtifactResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def GetFutureArtifact(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/datacatalog.DataCatalog/GetFutureArtifact',
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetArtifactRequest.SerializeToString,
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetArtifactResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def UpdateFutureArtifact(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/datacatalog.DataCatalog/UpdateFutureArtifact',
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.UpdateArtifactRequest.SerializeToString,
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.UpdateArtifactResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def AddTag(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/datacatalog.DataCatalog/AddTag',
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.AddTagRequest.SerializeToString,
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.AddTagResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def ListArtifacts(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/datacatalog.DataCatalog/ListArtifacts',
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.ListArtifactsRequest.SerializeToString,
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.ListArtifactsResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def ListDatasets(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/datacatalog.DataCatalog/ListDatasets',
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.ListDatasetsRequest.SerializeToString,
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.ListDatasetsResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def UpdateArtifact(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/datacatalog.DataCatalog/UpdateArtifact',
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.UpdateArtifactRequest.SerializeToString,
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.UpdateArtifactResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def GetOrExtendReservation(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/datacatalog.DataCatalog/GetOrExtendReservation',
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetOrExtendReservationRequest.SerializeToString,
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.GetOrExtendReservationResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def ReleaseReservation(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/datacatalog.DataCatalog/ReleaseReservation',
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.ReleaseReservationRequest.SerializeToString,
            flyteidl_dot_datacatalog_dot_datacatalog__pb2.ReleaseReservationResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)
