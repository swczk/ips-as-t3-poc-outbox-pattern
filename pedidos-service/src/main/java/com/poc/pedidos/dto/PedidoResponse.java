package com.poc.pedidos.dto;

import com.poc.pedidos.entity.Pedido;
import java.math.BigDecimal;
import java.time.OffsetDateTime;
import java.util.UUID;

public class PedidoResponse {

    private UUID id;
    private UUID clienteId;
    private BigDecimal valor;
    private String estado;
    private OffsetDateTime criadoEm;

    public static PedidoResponse from(Pedido pedido) {
        PedidoResponse r = new PedidoResponse();
        r.id        = pedido.getId();
        r.clienteId = pedido.getClienteId();
        r.valor     = pedido.getValor();
        r.estado    = pedido.getEstado();
        r.criadoEm  = pedido.getCriadoEm();
        return r;
    }

    public UUID getId() { return id; }
    public UUID getClienteId() { return clienteId; }
    public BigDecimal getValor() { return valor; }
    public String getEstado() { return estado; }
    public OffsetDateTime getCriadoEm() { return criadoEm; }
}