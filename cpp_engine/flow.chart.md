[mermaid.js]
```
    flowchart TD
    %% External Sources
    RTSP[RTSP IP Camera] -->|Network Stream| VI
    
    %% Ingestion Sub-section
    subgraph Ingestion["Video Ingestion Layer"]
        VI[VideoIngestion::startIngestion]
        Demux[av_read_frame]
        Route[routePacket]
        
        VI --> Demux
        Demux -->|AVPacket| Route
    end

    %% Memory & IPC Sub-section
    subgraph IPC["Inter-Process & Memory"]
        SHM[(/dev/shm Ring Buffer)]
        SQ[[SafeQueue]]
        
        Route -->|Zero-Copy| SHM
        Route -->|Clone Packet| SQ
    end
    
    SHM -.->|Read| GoLive[Go WebSocket Server]

    %% Storage Pipeline Sub-section
    subgraph Storage["Disk Writing Pipeline"]
        WW((writerWorker Thread))
        SR[SegmentRecorder]
        
        Norm[normalizeTimeline]
        Sanit[sanitizeTimestamps]
        Mux[av_interleaved_write_frame]
        
        SQ -->|Pop AVPacket| WW
        WW -->|Initialize / Rotate| SR
        WW -->|WritePacket| SR
        
        SR --> Norm
        Norm --> Sanit
        Sanit --> Mux
    end

    %% File System
    subgraph OS["Linux File System"]
        MKV[ .mkv Segment ]
    end

    Mux --> MKV
    
    %% Teardown Signals
    Shutdown((Destructor Signal)) -.->|Pushes nullptr| SQ
    WW -.->|Intercepts nullptr| Teardown[StopSegment]
```