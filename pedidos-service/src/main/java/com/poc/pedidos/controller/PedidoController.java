package com.poc.pedidos.controller;

import com.poc.pedidos.dto.CriarPedidoRequest;
import com.poc.pedidos.dto.PedidoResponse;
import com.poc.pedidos.service.PedidoService;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;
import java.util.UUID;

@RestController
public class PedidoController {

    private final PedidoService pedidoService;

    public PedidoController(PedidoService pedidoService) {
        this.pedidoService = pedidoService;
    }

    @PostMapping("/pedidos")
    @ResponseStatus(HttpStatus.CREATED)
    public PedidoResponse criarPedido(@RequestBody @Valid CriarPedidoRequest request) {
        return pedidoService.criarPedido(request);
    }

    @GetMapping("/pedidos/{id}")
    public PedidoResponse obterPedido(@PathVariable UUID id) {
        return pedidoService.obterPedido(id);
    }

    @GetMapping("/pedidos")
    public List<PedidoResponse> listarPedidosRecentes() {
        return pedidoService.listarPedidosRecentes();
    }

    @GetMapping("/health")
    public Map<String, String> health() {
        return Map.of("status", "ok");
    }
}