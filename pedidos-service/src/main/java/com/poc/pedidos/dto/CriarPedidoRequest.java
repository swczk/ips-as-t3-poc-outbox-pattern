package com.poc.pedidos.dto;

import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.NotNull;

import java.math.BigDecimal;
import java.util.UUID;

public class CriarPedidoRequest {

    @NotNull(message = "clienteId é obrigatório")
    private UUID clienteId;

    @NotNull(message = "valor é obrigatório")
    @DecimalMin(value = "0.01", message = "valor tem de ser maior do que zero")
    private BigDecimal valor;

    public UUID getClienteId() { return clienteId; }
    public void setClienteId(UUID clienteId) { this.clienteId = clienteId; }

    public BigDecimal getValor() { return valor; }
    public void setValor(BigDecimal valor) { this.valor = valor; }
}