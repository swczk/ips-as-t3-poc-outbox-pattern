package com.poc.pedidos.controller;

import com.poc.pedidos.dto.CriarPedidoRequest;
import com.poc.pedidos.dto.PedidoResponse;
import com.poc.pedidos.service.PedidoService;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
public class PedidoController {

    private final PedidoService pedidoService;

    public PedidoController(PedidoService pedidoService) {
        this.pedidoService = pedidoService;
    }

    @PostMapping("/pedidos")
    @ResponseStatus(HttpStatus.CREATED)
    public PedidoResponse criarPedido(@RequestBody @Valid CriarPedidoRequest request)
            throws Exception {
        return pedidoService.criarPedido(request);
    }

    @GetMapping("/health")
    public Map<String, String> health() {
        return Map.of("status", "ok");
    }
}
