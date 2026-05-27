package com.poc.pedidos.entity;

import jakarta.persistence.*;
import java.time.OffsetDateTime;
import java.util.UUID;

@Entity
@Table(name = "outbox")
public class OutboxEvent {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    @Column(nullable = false)
    private String tipo;

    @Column(nullable = false, columnDefinition = "jsonb")
    private String payload;

    @Column(nullable = false)
    private Boolean publicado = false;

    @Column(name = "criado_em")
    private OffsetDateTime criadoEm = OffsetDateTime.now();

    @Column(nullable = false)
    private Integer tentativas = 0;

    public UUID getId() { return id; }
    public String getTipo() { return tipo; }
    public void setTipo(String tipo) { this.tipo = tipo; }
    public String getPayload() { return payload; }
    public void setPayload(String payload) { this.payload = payload; }
    public Boolean getPublicado() { return publicado; }
    public OffsetDateTime getCriadoEm() { return criadoEm; }
    public Integer getTentativas() { return tentativas; }
}
